package pkg

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
)

type DiffResult struct {
	Add map[schema.GroupVersionKind][]Matcher
	Remove map[schema.GroupVersionKind][]Matcher
}

var empty  = struct{}{}

func CompareMatchers(existing, latest map[schema.GroupVersionKind][]Matcher) DiffResult {
	var result DiffResult

	if existing == nil {
		result = DiffResult{
			Add: latest,
		}
	} else if latest == nil {
		result = DiffResult{
			Remove: existing,
		}
	} else {
		result =DiffResult{
			Add:    map[schema.GroupVersionKind][]Matcher{},
			Remove: map[schema.GroupVersionKind][]Matcher{},
		}

		for gvk, matchers := range latest {
			existingMatchers, ok := existing[gvk]
			if !ok {
				result.Add[gvk] = matchers
			} else {
				mm := map[Matcher]struct{}{}
				for _, entry := range existingMatchers {
					mm[entry] = empty
				}
				for _, m := range matchers {
					if _, ok := mm[m]; ok {
						delete(mm, m)
					} else {
						result.Add[gvk] = append(result.Add[gvk], m)
					}
					for m := range mm {
						result.Remove[gvk] = append(result.Add[gvk], m)
					}
				}
			}
		}
		for gvk, matchers := range existing {
			if _, ok := latest[gvk]; !ok {
				result.Remove[gvk] = matchers
			}
		}
	}
	return result
}

func UpdateMatchers(
	// Module -> Matchers
	moduleToMatchers map[types.NamespacedName]map[schema.GroupVersionKind][]Matcher,
	// Kind -> Matchers -> []Module
	kindToModule map[schema.GroupVersionKind]map[Matcher][]types.NamespacedName,
	resource types.NamespacedName,
	matchers map[schema.GroupVersionKind][]Matcher,
	result DiffResult,
) {
	moduleToMatchers[resource] = matchers

	for gvk, mlist := range result.Add {
		m2 := kindToModule[gvk]
		if m2 == nil {
			m2 = map[Matcher][]types.NamespacedName{}
		}
		for _, m := range mlist {
			rlist := m2[m] // pointer ?
			for _, e := range rlist {
				if reflect.DeepEqual(e,)
			}

		}

		kindToModule[gvk] = m2
	}

	for gvk, mlist := range result.Remove {

	}


}
