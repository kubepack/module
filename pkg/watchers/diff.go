package watchers

import (
	"gomodules.xyz/sets"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type DiffResult struct {
	Add    map[schema.GroupVersionKind]sets.Uint64
	Remove map[schema.GroupVersionKind]sets.Uint64
}

func (r DiffResult) Empty() bool {
	return len(r.Add)+len(r.Remove) == 0
}

func (w *ModuleWatchers) compareMatchers(module types.NamespacedName, latest map[schema.GroupVersionKind]sets.Uint64) DiffResult {
	result := DiffResult{
		Add:    map[schema.GroupVersionKind]sets.Uint64{},
		Remove: map[schema.GroupVersionKind]sets.Uint64{},
	}

	existing := w.ModuleToMatchers[module]

	if len(existing) == 0 {
		result.Add = latest
	} else if len(latest) == 0 {
		result.Remove = existing
	} else {
		for gvk, latestIdSet := range latest {
			result.Add[gvk] = latestIdSet.Difference(existing[gvk])
			result.Remove[gvk] = existing[gvk].Difference(latestIdSet)
		}
		for gvk, idSet := range existing {
			if _, ok := latest[gvk]; !ok {
				result.Remove[gvk] = idSet
			}
		}
	}
	return result
}
