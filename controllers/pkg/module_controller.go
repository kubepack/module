/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pkg

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"kmodules.xyz/client-go/discovery"
	"kubepack.dev/lib-helm/pkg/repo"
	pkgapi "kubepack.dev/module/apis/pkg/v1alpha1"
	"kubepack.dev/module/pkg"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sync"
)

// ModuleReconciler reconciles a Module object
type ModuleReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	DC           dynamic.Interface
	ClientGetter genericclioptions.RESTClientGetter
	Mapper       discovery.ResourceMapper

	ChartRegistry repo.IRegistry

	Mgr  ctrl.Manager
	ctrl controller.Controller

	mu sync.Mutex

	// Module -> Matchers
	ModuleToMatchers map[types.NamespacedName]map[schema.GroupVersionKind][]pkg.Matcher

	// Kind -> Matchers -> []Module
	KindToModule map[schema.GroupVersionKind]map[pkg.Matcher][]types.NamespacedName
}

//+kubebuilder:rbac:groups=pkgapi.kubepack.com,resources=modules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pkgapi.kubepack.com,resources=modules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pkgapi.kubepack.com,resources=modules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Module object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ModuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// your logic here
	var module pkgapi.Module
	if err := r.Get(ctx, req.NamespacedName, &module); err != nil {
		log.Error(err, "unable to fetch Module")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// FIX(tamal): Observed generation checking is not enough, since the overridden values can change.

	matchers := map[schema.GroupVersionKind][]pkg.Matcher{}
	for _, action := range module.Spec.Actions {
		runner := pkg.ActionRunner{
			DC:            r.DC,
			ClientGetter:  r.ClientGetter,
			Mapper:        r.Mapper,
			ModuleName:    module.Name,
			Namespace:     module.Namespace,
			Action:        action,
			EdgeList:      module.Spec.EdgeList,
			ChartRegistry: r.ChartRegistry,
			Matchers:      matchers,
		}
		err := runner.Execute()
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	r.mu.Lock()
	pkg.UpdateMatchers(
		r.ModuleToMatchers,
		r.KindToModule,
		req.NamespacedName,
		pkg.CompareMatchers(r.ModuleToMatchers[req.NamespacedName], matchers),
	)
	r.mu.Unlock()

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModuleReconciler) SetupWithManager(mgr ctrl.Manager) (err error) {
	r.ctrl, err = ctrl.NewControllerManagedBy(mgr).
		For(&pkgapi.Module{}).
		Build(r)
	return err
}

func (r *ModuleReconciler) ResourceToModules(ctx context.Context) handler.MapFunc {
	log := ctrl.LoggerFrom(ctx)
	return func(o client.Object) []ctrl.Request {
		kinds, unversioned, err := r.Mgr.GetScheme().ObjectKinds(o)
		if err != nil {
			log.Error(err, "failed to detect kind")
			return nil
		}
		if unversioned {
			log.Info("object is unversioned", "type", reflect.TypeOf(o))
			return nil
		}

		result := []ctrl.Request{}

		// TODO(tamal): Take lock on KindToModule?

		for _, gvk := range kinds {
			matchers, ok := r.KindToModule[gvk]
			if !ok {
				continue
			}
			for matcher, modules := range matchers {
				if matcher.Matches(o) {
					for _, module := range modules {
						result = append(result, ctrl.Request{NamespacedName: module})
					}
				}
			}
		}
		return result
	}
}
