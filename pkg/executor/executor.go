package executor

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"kubepack.dev/lib-helm/pkg/repo"
	"kubepack.dev/module/pkg/api"
	"strings"
	"text/template"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/discovery"
	dynamicfactory "kmodules.xyz/client-go/dynamic/factory"
	rsapi "kmodules.xyz/resource-metadata/apis/meta/v1alpha1"
	"kmodules.xyz/resource-metadata/pkg/graph"
	"kmodules.xyz/resource-metadata/pkg/tableconvertor"
	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/pkg/lib"
	"kubepack.dev/lib-helm/pkg/action"
	"kubepack.dev/lib-helm/pkg/engine"
	"kubepack.dev/lib-helm/pkg/values"
	pkgapi "kubepack.dev/module/apis/pkg/v1alpha1"
)

var ModuleStore = map[string]*engine.State{}

type ActionExecutor struct {
	DC           dynamic.Interface
	ClientGetter genericclioptions.RESTClientGetter
	Mapper       discovery.ResourceMapper

	ModuleName string
	Namespace  string
	Action     pkgapi.Action
	EdgeList   []rsapi.NamedEdge
	err        error

	ChartRegistry repo.IRegistry

	Matchers map[schema.GroupVersionKind][]api.Matcher
}

func (a *ActionExecutor) Execute() error {
	if a.MeetsPrerequisites() {
		a.Apply().WaitUntilReady()
		if a.Err() != nil {
			// if ae, ok := a.Err().(AlreadyErrored); ok {
			// Action was already errored out, Rest before reusing
		}
	} else {
		if a.Err() != nil {
			// if ae, ok := a.Err().(AlreadyErrored); ok {
			// Action was already errored out, Rest before reusing
		}
	}
	return nil
}

// Do we need this?
func (a *ActionExecutor) ResetError() {
	a.err = nil
}

func (a *ActionExecutor) Err() error {
	return a.err
}

func (a *ActionExecutor) MeetsPrerequisites() bool {
	if a.err != nil {
		a.err = NewAlreadyErrored(a.err)
		return false
	}

	if a.ModuleName == "" {
		a.err = fmt.Errorf("flow name is required")
		return false
	}

	return a.resourceExists(context.TODO(), a.Action.Prerequisites.RequiredResources)
}

func (a *ActionExecutor) Apply() *ActionExecutor {
	if a.err != nil {
		a.err = NewAlreadyErrored(a.err)
		return a
	}

	chrt, err := a.ChartRegistry.GetChart(a.Action.ChartRef.URL, a.Action.Name, a.Action.ChartRef.Version)
	if err != nil {
		a.err = err
		return a
	}

	opts := values.Options{
		// ReplaceValues: nil,
		ValuesFile:   a.Action.ValuesFile,
		ValuesPatch:  a.Action.ValuesPatch,
		StringValues: nil,
		Values:       nil,
		KVPairs:      nil,
	}

	finder := graph.ObjectFinder{
		Factory: dynamicfactory.New(a.DC),
		Mapper:  a.Mapper,
	}

	for _, o := range a.Action.ValueOverrides {
		var selector *metav1.LabelSelector
		var name string
		var obj *unstructured.Unstructured

		if o.ObjRef != nil {
			selector = o.ObjRef.Src.Selector
			name = o.ObjRef.Src.Name

			if o.ObjRef.Src.UseAction != "" && name == "" {
				state, ok := ModuleStore[o.ObjRef.Src.UseAction]
				if !ok {
					a.err = fmt.Errorf("can't find flow state for release %s", o.ObjRef.Src.UseAction)
					return a
				}
				// initialize engine if needed
				err = state.Init()
				if err != nil {
					a.err = fmt.Errorf("failed to initialize engine, reason; %v", err)
					klog.Errorln(err)
					return a
				}

				tpl := o.ObjRef.Src.NameTemplate
				if tpl != "" {
					l := TemplateList{}
					l.Add(tpl)
					result, err := state.Engine.Render(l)
					if err != nil {
						a.err = err
						return a
					}
					name = TemplateList(result).Get(tpl)
				} else if o.ObjRef.Src.Selector != nil {
					l := TemplateList{}
					for _, v := range o.ObjRef.Src.Selector.MatchLabels {
						l.Add(v)
					}
					for _, expr := range o.ObjRef.Src.Selector.MatchExpressions {
						for _, v := range expr.Values {
							l.Add(v)
						}
					}

					l, err = state.Engine.Render(l)
					if err != nil {
						a.err = err
						return a
					}

					var sel metav1.LabelSelector
					if o.ObjRef.Src.Selector.MatchLabels != nil {
						sel.MatchLabels = make(map[string]string)
					}
					for k, v := range o.ObjRef.Src.Selector.MatchLabels {
						sel.MatchLabels[k] = l.Get(v)
					}
					if len(o.ObjRef.Src.Selector.MatchExpressions) > 0 {
						sel.MatchExpressions = make([]metav1.LabelSelectorRequirement, 0, len(o.ObjRef.Src.Selector.MatchExpressions))
					}
					for _, expr := range o.ObjRef.Src.Selector.MatchExpressions {
						ne := expr
						ne.Values = make([]string, 0, len(expr.Values))
						for _, v := range expr.Values {
							ne.Values = append(ne.Values, l.Get(v))
						}
						sel.MatchExpressions = append(sel.MatchExpressions, ne)
					}
					selector = &sel
				}
			}

			gvk := schema.FromAPIVersionAndKind(o.ObjRef.Src.Target.APIVersion, o.ObjRef.Src.Target.Kind)

			m := api.Matcher{
				Name:      name,
				Namespace: a.Namespace,
				Selector:  selector,
				// owner ?
			}
			a.Matchers[gvk] = append(a.Matchers[gvk], m)
			obj, err = finder.Locate(&rsapi.ObjectLocator{
				Src: rsapi.ObjectRef{
					Target:    o.ObjRef.Src.Target,
					Selector:  selector,
					Name:      name,
					Namespace: a.Namespace,
				},
				Path: o.ObjRef.Paths,
			}, a.EdgeList)
			if err != nil {
				a.err = err
				return a
			}
			fmt.Println(obj.GetName())
		}

		var buf bytes.Buffer
		for _, kv := range o.Values {
			//kv.Type
			//kv.Format
			//kv.Key
			//kv.FieldPath
			//kv.FieldPathTemplate
			if kv.ValueRef.Raw != "" {
				switch kv.ValueRef.Type {
				case "string":
					opts.StringValues = append(opts.StringValues, fmt.Sprintf("%s=%v", kv.Key, kv.ValueRef.Raw))
				case "nil", "null":
					// See https://helm.sh/docs/chart_template_guide/values_files/#deleting-a-default-key
					opts.Values = append(opts.Values, fmt.Sprintf("%s=null", kv.Key))
				default:
					opts.Values = append(opts.Values, fmt.Sprintf("%s=%v", kv.Key, kv.ValueRef.Raw))
				}
			} else if kv.ValueRef.FieldPathTemplate != "" {
				if obj == nil {
					a.err = fmt.Errorf("failed to locate object for Action %s", a.Action.Name)
					return a
				}

				tpl, err := template.New("").Funcs(tableconvertor.TxtFuncMap()).Parse(kv.ValueRef.FieldPathTemplate)
				if err != nil {
					a.err = fmt.Errorf("failed to parse path template %s, reason: %v", kv.ValueRef.FieldPathTemplate, err)
					return a
				}
				buf.Reset()
				err = tpl.Execute(&buf, obj.UnstructuredContent())
				if err != nil {
					a.err = fmt.Errorf("failed to resolve path template %s, reason: %v", kv.ValueRef.FieldPathTemplate, err)
					return a
				}
				switch kv.ValueRef.Type {
				case "string":
					opts.StringValues = append(opts.StringValues, fmt.Sprintf("%s=%v", kv.Key, buf.String()))
				case "nil", "null":
					// See https://helm.sh/docs/chart_template_guide/values_files/#deleting-a-default-key
					opts.Values = append(opts.Values, fmt.Sprintf("%s=null", kv.Key))
				default:
					opts.Values = append(opts.Values, fmt.Sprintf("%s=%v", kv.Key, buf.String()))
				}
			} else if kv.ValueRef.FieldPath != "" {
				if obj == nil {
					a.err = fmt.Errorf("failed to locate object for Action %s", a.Action.Name)
					return a
				}

				path := strings.Trim(kv.ValueRef.FieldPath, ".")
				v, ok, err := unstructured.NestedFieldNoCopy(obj.UnstructuredContent(), strings.Split(path, ".")...)
				if err != nil {
					a.err = err
					return a
				}
				if !ok {
					// this is the standard behavior Helm template follows
					v = ""
				}
				opts.KVPairs = append(opts.KVPairs, values.KV{
					K: kv.Key,
					V: v,
				})
			}
		}
	}

	// Now pass this as replace values
	vals, err := opts.MergeValues(chrt.Chart)
	if err != nil {
		a.err = err
		return a
	}

	deployer, err := action.NewDeployer(a.ClientGetter, a.Namespace, "storage.x-helm.dev/apps")
	if err != nil {
		a.err = err
		return a
	}
	deployer.WithRegistry(lib.DefaultRegistry).WithOptions(action.DeployOptions{
		ChartURL:  a.Action.ChartRef.URL,
		ChartName: a.Action.ChartRef.Name,
		Version:   a.Action.ChartRef.Version,
		Values: values.Options{
			ReplaceValues: vals,
		},
		DryRun:                   false,
		DisableHooks:             false,
		Replace:                  false,
		Wait:                     true,
		Devel:                    false,
		Timeout:                  15 * time.Minute,
		Namespace:                a.Namespace,
		ReleaseName:              a.Action.Name,
		Description:              "Deploy Module",
		Atomic:                   false,
		SkipCRDs:                 true,
		SubNotes:                 false,
		DisableOpenAPIValidation: false,
		IncludeCRDs:              false,
		PartOf:                   a.ModuleName,
		CreateNamespace:          true,
		Force:                    false,
		Recreate:                 false,
		CleanupOnFail:            false,
	})
	_, s2, err := deployer.Run()
	if err != nil {
		a.err = err
	}
	ModuleStore[s2.ReleaseName] = s2

	return a
}

func (a *ActionExecutor) WaitUntilReady() {
	if a.err != nil {
		a.err = NewAlreadyErrored(a.err)
		return
	}

	if a.Action.ReadinessCriteria.Timeout.Duration == 0 {
		a.Action.ReadinessCriteria.Timeout = metav1.Duration{Duration: 15 * time.Minute}
	}
	// start := time.Now()
	// calculate timeout

	ctx, cancel := context.WithTimeout(context.TODO(), a.Action.ReadinessCriteria.Timeout.Duration)
	defer cancel()

	rready := a.resourceExists(ctx, a.Action.Prerequisites.RequiredResources)
	if a.err != nil {
		return
	}
	if !rready {
		return
	}

	// WaitForFlags
	waitflags := make([]v1alpha1.WaitFlags, 0, len(a.Action.ReadinessCriteria.WaitFors))
	for _, w := range a.Action.ReadinessCriteria.WaitFors {
		waitflags = append(waitflags, v1alpha1.WaitFlags{
			Resource:     w.Resource,
			Labels:       w.Labels,
			All:          w.All,
			Timeout:      w.Timeout,
			ForCondition: w.ForCondition,
		})
	}

	var buf bytes.Buffer
	printer := lib.WaitForPrinter{
		Name:      a.Action.Name,
		Namespace: a.Namespace,
		WaitFors:  waitflags,
		W:         &buf,
	}
	err := printer.Do()
	if err != nil {
		a.err = err
		return
	}
	if buf.Len() > 0 {
		klog.Infoln("running commands:")
		for _, line := range strings.Split(buf.String(), "\n") {
			klog.Infoln(line)
		}
	}

	checker := lib.WaitForChecker{
		Namespace:    a.Namespace,
		WaitFors:     waitflags,
		ClientGetter: a.ClientGetter,
	}
	err = checker.Do()
	if err != nil {
		a.err = err
		return
	}
}

func (a *ActionExecutor) resourceExists(ctx context.Context, resources []metav1.GroupVersionResource) bool {
	for _, r := range resources {
		exists, err := IsResourceExistsAndReady(ctx, a.DC, a.Mapper, schema.GroupVersionResource{
			Group:    r.Group,
			Version:  r.Version,
			Resource: r.Resource,
		})
		if err != nil {
			a.err = err
			return false
		}
		if !exists {
			return false
		}
	}
	return true
}

func (a *ActionExecutor) IsReady() bool {
	if a.err != nil {
		a.err = NewAlreadyErrored(a.err)
		return false
	}

	return false
}

type AlreadyErrored struct {
	underlying error
}

func NewAlreadyErrored(underlying error) error {
	if _, ok := underlying.(AlreadyErrored); ok {
		return underlying
	}
	return AlreadyErrored{underlying: underlying}
}

func (e AlreadyErrored) Error() string {
	return e.underlying.Error()
}

func (e AlreadyErrored) Underlying() error {
	return e.underlying
}

type TemplateList map[string]string

func (l TemplateList) Add(tpl string) {
	l[base64.URLEncoding.EncodeToString([]byte(tpl))] = tpl
}

func (l TemplateList) Get(tpl string) string {
	return l[base64.URLEncoding.EncodeToString([]byte(tpl))]
}
