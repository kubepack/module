package watchers

import (
	"context"
	"reflect"
	"sync"

	"gomodules.xyz/sets"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ksets "kmodules.xyz/sets"
	"kubepack.dev/module/apis/pkg/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type ModuleWatchers struct {
	lock sync.Mutex

	// WARNING: How to handle version skew
	// example: different locators depend on different version
	// possible answer, matchers depend on object metadata, so same for every version
	Watchers ksets.GroupKind

	// map_index -> Matcher
	Matchers map[uint64]*v1alpha1.Matcher

	// Module -> Matchers
	ModuleToMatchers map[types.NamespacedName]map[schema.GroupVersionKind]sets.Uint64

	// Kind -> Matchers -> []Module
	KindToModule map[schema.GroupVersionKind]map[uint64]ksets.NamespacedName
}

func New() *ModuleWatchers {
	return &ModuleWatchers{
		Watchers:         ksets.NewGroupKind(),
		Matchers:         map[uint64]*v1alpha1.Matcher{},
		ModuleToMatchers: map[types.NamespacedName]map[schema.GroupVersionKind]sets.Uint64{},
		KindToModule:     map[schema.GroupVersionKind]map[uint64]ksets.NamespacedName{},
	}
}

func (w *ModuleWatchers) ResourceToModules(ctx context.Context, mgr ctrl.Manager) handler.MapFunc {
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

		w.lock.Lock()
		defer w.lock.Unlock()

		result := []ctrl.Request{}
		for _, gvk := range kinds {
			matchers, ok := w.KindToModule[gvk]
			if !ok {
				continue
			}
			for midx, modules := range matchers {
				if w.Matchers[midx].Matches(o) {
					for module := range modules {
						result = append(result, ctrl.Request{NamespacedName: module})
					}
				}
			}
		}
		return result
	}
}

func (w *ModuleWatchers) UpdateMatchers(
	ctx context.Context,
	mgr ctrl.Manager,
	ctrl2 controller.Controller,
	module types.NamespacedName,
	matchers map[schema.GroupVersionKind][]v1alpha1.Matcher) []v1alpha1.Watcher {

	w.lock.Lock()
	defer w.lock.Unlock()

	latest := map[schema.GroupVersionKind]sets.Uint64{}
	for gvk, mm := range matchers {
		idSet := make(sets.Uint64, len(mm))
		for i, m := range mm {
			midx := m.MapIndex()
			idSet.Insert(midx)
			w.Matchers[midx] = &mm[i] // WARN: updates Matchers
		}
		latest[gvk] = idSet
	}

	diff := w.compareMatchers(module, latest)
	if diff.Empty() {
		return nil
	}

	w.ModuleToMatchers[module] = latest

	for gvk, idSet := range diff.Add {
		existing := w.KindToModule[gvk]
		if existing == nil {
			existing = map[uint64]ksets.NamespacedName{}
		}
		for id := range idSet {
			existing[id] = ksets.NewNamespacedName(module).Union(existing[id])
		}
		w.KindToModule[gvk] = existing
	}
	for gvk, idSet := range diff.Remove {
		existing := w.KindToModule[gvk]
		if existing == nil {
			continue
		}
		for id := range idSet {
			if moduleSet, ok := existing[id]; ok {
				moduleSet.Delete(module)
			}
		}
	}

	log := ctrl.LoggerFrom(ctx)
	for gvk := range matchers {
		if !w.Watchers.Has(gvk.GroupKind()) {
			obj, err := mgr.GetScheme().New(gvk)
			if err != nil {
				log.Error(err, "failed to create object", "gvk", gvk)
				continue
			}
			if err := ctrl2.Watch(
				&source.Kind{Type: obj.(client.Object)},
				handler.EnqueueRequestsFromMapFunc(w.ResourceToModules(ctx, mgr)),
			); err != nil {
				log.Error(err, "failed to set up watcher", "gvk", gvk)
				continue
			}

			w.Watchers.Insert(gvk.GroupKind())
		}
	}

	watchers := make([]v1alpha1.Watcher, 0, len(latest))
	for gvk, idSet := range latest {
		matchers := make([]v1alpha1.Matcher, 0, idSet.Len())
		for midx := range idSet {
			matchers = append(matchers, *w.Matchers[midx])
		}
		watchers = append(watchers, v1alpha1.Watcher{
			Kind: v1alpha1.GroupVersionKind{
				Group:   gvk.Group,
				Version: gvk.Version,
				Kind:    gvk.Kind,
			},
			Matchers: matchers,
		})
	}
	return watchers
}

func (w *ModuleWatchers) DeleteModule(module types.NamespacedName) {
	w.lock.Lock()
	defer w.lock.Unlock()

	for gvk, idSet := range w.ModuleToMatchers[module] {
		modulesByIds := w.KindToModule[gvk]
		for idx := range idSet {
			modulesByIds[idx].Delete(module)
		}
	}
	delete(w.ModuleToMatchers, module)
}
