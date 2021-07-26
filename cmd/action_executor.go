package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
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

type ActionRunner struct {
	dc           dynamic.Interface
	ClientGetter genericclioptions.RESTClientGetter
	mapper       discovery.ResourceMapper

	ModuleName string
	Namespace  string
	action     pkgapi.Action
	EdgeList   []rsapi.NamedEdge
	err        error
}

func (runner *ActionRunner) Execute() error {
	if runner.MeetsPrerequisites() {
		runner.Apply().WaitUntilReady()
		if runner.Err() != nil {
			// if ae, ok := runner.Err().(AlreadyErrored); ok {
			// action was already errored out, Rest before reusing
		}
	} else {
		if runner.Err() != nil {
			// if ae, ok := runner.Err().(AlreadyErrored); ok {
			// action was already errored out, Rest before reusing
		}
	}
	return nil
}

// Do we need this?
func (e *ActionRunner) ResetError() {
	e.err = nil
}

func (e *ActionRunner) Err() error {
	return e.err
}

func (e *ActionRunner) MeetsPrerequisites() bool {
	if e.err != nil {
		e.err = NewAlreadyErrored(e.err)
		return false
	}

	if e.ModuleName == "" {
		e.err = fmt.Errorf("flow name is required")
		return false
	}

	return e.resourceExists(context.TODO(), e.action.Prerequisites.RequiredResources)
}

func (e *ActionRunner) Apply() *ActionRunner {
	if e.err != nil {
		e.err = NewAlreadyErrored(e.err)
		return e
	}

	chrt, err := lib.DefaultRegistry.GetChart(e.action.ChartRef.URL, e.action.Name, e.action.ChartRef.Version)
	if err != nil {
		e.err = err
		return e
	}

	opts := values.Options{
		// ReplaceValues: nil,
		ValuesFile:   e.action.ValuesFile,
		ValuesPatch:  e.action.ValuesPatch,
		StringValues: nil,
		Values:       nil,
		KVPairs:      nil,
	}

	finder := graph.ObjectFinder{
		Factory: dynamicfactory.New(e.dc),
		Mapper:  e.mapper,
	}

	for _, o := range e.action.ValueOverrides {
		var selector *metav1.LabelSelector
		var name string
		var obj *unstructured.Unstructured

		if o.ObjRef != nil {
			selector = o.ObjRef.Src.Selector
			name = o.ObjRef.Src.Name

			if o.ObjRef.Src.UseAction != "" && name == "" {
				state, ok := ModuleStore[o.ObjRef.Src.UseAction]
				if !ok {
					e.err = fmt.Errorf("can't find flow state for release %s", o.ObjRef.Src.UseAction)
					return e
				}
				// initialize engine if needed
				err = state.Init()
				if err != nil {
					e.err = fmt.Errorf("failed to initialize engine, reason; %v", err)
					klog.Errorln(err)
					return e
				}

				tpl := o.ObjRef.Src.NameTemplate
				if tpl != "" {
					l := TemplateList{}
					l.Add(tpl)
					result, err := state.Engine.Render(l)
					if err != nil {
						e.err = err
						return e
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
						e.err = err
						return e
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

			obj, err = finder.Locate(&rsapi.ObjectLocator{
				Src: rsapi.ObjectRef{
					Target:    o.ObjRef.Src.Target,
					Selector:  selector,
					Name:      name,
					Namespace: e.Namespace,
				},
				Path: o.ObjRef.Paths,
			}, e.EdgeList)
			if err != nil {
				e.err = err
				return e
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
					e.err = fmt.Errorf("failed to locate object for action %s", e.action.Name)
					return e
				}

				tpl, err := template.New("").Funcs(tableconvertor.TxtFuncMap()).Parse(kv.ValueRef.FieldPathTemplate)
				if err != nil {
					e.err = fmt.Errorf("failed to parse path template %s, reason: %v", kv.ValueRef.FieldPathTemplate, err)
					return e
				}
				buf.Reset()
				err = tpl.Execute(&buf, obj.UnstructuredContent())
				if err != nil {
					e.err = fmt.Errorf("failed to resolve path template %s, reason: %v", kv.ValueRef.FieldPathTemplate, err)
					return e
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
					e.err = fmt.Errorf("failed to locate object for action %s", e.action.Name)
					return e
				}

				path := strings.Trim(kv.ValueRef.FieldPath, ".")
				v, ok, err := unstructured.NestedFieldNoCopy(obj.UnstructuredContent(), strings.Split(path, ".")...)
				if err != nil {
					e.err = err
					return e
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
		e.err = err
		return e
	}

	deployer, err := action.NewDeployer(e.ClientGetter, e.Namespace, "storage.x-helm.dev/apps")
	if err != nil {
		e.err = err
		return e
	}
	deployer.WithRegistry(lib.DefaultRegistry).WithOptions(action.DeployOptions{
		ChartURL:  e.action.ChartRef.URL,
		ChartName: e.action.ChartRef.Name,
		Version:   e.action.ChartRef.Version,
		Values: values.Options{
			ReplaceValues: vals,
		},
		DryRun:                   false,
		DisableHooks:             false,
		Replace:                  false,
		Wait:                     true,
		Devel:                    false,
		Timeout:                  15 * time.Minute,
		Namespace:                e.Namespace,
		ReleaseName:              e.action.Name,
		Description:              "Deploy Module",
		Atomic:                   false,
		SkipCRDs:                 true,
		SubNotes:                 false,
		DisableOpenAPIValidation: false,
		IncludeCRDs:              false,
		PartOf:                   e.ModuleName,
		CreateNamespace:          true,
		Force:                    false,
		Recreate:                 false,
		CleanupOnFail:            false,
	})
	_, s2, err := deployer.Run()
	if err != nil {
		e.err = err
	}
	ModuleStore[s2.ReleaseName] = s2

	return e
}

func (e *ActionRunner) WaitUntilReady() {
	if e.err != nil {
		e.err = NewAlreadyErrored(e.err)
		return
	}

	if e.action.ReadinessCriteria.Timeout.Duration == 0 {
		e.action.ReadinessCriteria.Timeout = metav1.Duration{Duration: 15 * time.Minute}
	}
	// start := time.Now()
	// calculate timeout

	ctx, cancel := context.WithTimeout(context.TODO(), e.action.ReadinessCriteria.Timeout.Duration)
	defer cancel()

	rready := e.resourceExists(ctx, e.action.Prerequisites.RequiredResources)
	if e.err != nil {
		return
	}
	if !rready {
		return
	}

	// WaitForFlags
	waitflags := make([]v1alpha1.WaitFlags, 0, len(e.action.ReadinessCriteria.WaitFors))
	for _, w := range e.action.ReadinessCriteria.WaitFors {
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
		Name:      e.action.Name,
		Namespace: e.Namespace,
		WaitFors:  waitflags,
		W:         &buf,
	}
	err := printer.Do()
	if err != nil {
		e.err = err
		return
	}
	if buf.Len() > 0 {
		klog.Infoln("running commands:")
		for _, line := range strings.Split(buf.String(), "\n") {
			klog.Infoln(line)
		}
	}

	checker := lib.WaitForChecker{
		Namespace:    e.Namespace,
		WaitFors:     waitflags,
		ClientGetter: e.ClientGetter,
	}
	err = checker.Do()
	if err != nil {
		e.err = err
		return
	}
}

func (e *ActionRunner) resourceExists(ctx context.Context, resources []metav1.GroupVersionResource) bool {
	for _, r := range resources {
		exists, err := IsResourceExistsAndReady(ctx, e.dc, e.mapper, schema.GroupVersionResource{
			Group:    r.Group,
			Version:  r.Version,
			Resource: r.Resource,
		})
		if err != nil {
			e.err = err
			return false
		}
		if !exists {
			return false
		}
	}
	return true
}

func (e *ActionRunner) IsReady() bool {
	if e.err != nil {
		e.err = NewAlreadyErrored(e.err)
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
