package lib

import (
	"context"
	"reflect"
	"sync"

	"gomodules.xyz/sets"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"kubepack.dev/module/pkg/api"
	sets2 "kubepack.dev/module/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

type ModuleWatchers struct {
	lock sync.Mutex

	// map_index -> Matcher
	Matchers map[uint64]*api.Matcher

	// Module -> Matchers
	ModuleToMatchers map[types.NamespacedName]map[schema.GroupVersionKind]sets.Uint64

	// Kind -> Matchers -> []Module
	KindToModule map[schema.GroupVersionKind]map[uint64]sets2.NamespacedName
}

func New() *ModuleWatchers {
	return &ModuleWatchers{
		Matchers:         map[uint64]*api.Matcher{},
		ModuleToMatchers: map[types.NamespacedName]map[schema.GroupVersionKind]sets.Uint64{},
		KindToModule:     map[schema.GroupVersionKind]map[uint64]sets2.NamespacedName{},
	}
}

func (r *ModuleWatchers) ResourceToModules(ctx context.Context, mgr ctrl.Manager) handler.MapFunc {
	log := ctrl.LoggerFrom(ctx)
	return func(o client.Object) []ctrl.Request {
		kinds, unversioned, err := mgr.GetScheme().ObjectKinds(o)
		if err != nil {
			log.Error(err, "failed to detect kind")
			return nil
		}
		if unversioned {
			log.Info("object is unversioned", "type", reflect.TypeOf(o))
			return nil
		}

		r.lock.Lock()
		defer r.lock.Unlock()

		result := []ctrl.Request{}
		for _, gvk := range kinds {
			matchers, ok := r.KindToModule[gvk]
			if !ok {
				continue
			}
			for midx, modules := range matchers {
				if r.Matchers[midx].Matches(o) {
					for module := range modules {
						result = append(result, ctrl.Request{NamespacedName: module})
					}
				}
			}
		}
		return result
	}
}
