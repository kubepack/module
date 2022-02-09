package v1alpha1

import (
	"encoding/gob"
	"fmt"
	"github.com/zeebo/xxh3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Matcher struct {
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
	// +optional
	Owner *OwnerReference `json:"owner,omitempty"`
}

// OwnerReference contains enough information to let you identify an owning
// object. An owning object must be in the same namespace as the dependent, or
// be cluster-scoped, so there is no namespace field.
type OwnerReference struct {
	// API version of the referent.
	APIVersion string `json:"apiVersion"`
	// Kind of the referent.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	Kind string `json:"kind"`
	// Name of the referent.
	// More info: http://kubernetes.io/docs/user-guide/identifiers#names
	Name string `json:"name"`
	// If true, this reference points to the managing controller.
	// +optional
	Controller *bool `json:"controller,omitempty"`
}

func (m *Matcher) MapIndex() uint64 {
	if m == nil {
		return 0
	}
	h := xxh3.New()
	if err := gob.NewEncoder(h).Encode(m); err != nil {
		panic(fmt.Errorf("failed to gob encode %#v: %w", m, err))
	}
	return h.Sum64()
}

func (m Matcher) Compare(other Matcher) int {
	if m.Name < other.Name {
		return -1
	}
	if m.Name > other.Name {
		return +1
	}
	if m.Namespace < other.Namespace {
		return -1
	}
	if m.Namespace > other.Namespace {
		return +1
	}
	lhsIdx := m.MapIndex()
	rhsIdx := other.MapIndex()
	if lhsIdx < rhsIdx {
		return -1
	} else if lhsIdx > rhsIdx {
		return +1
	}
	return 0
}

func (m Matcher) Equal(other Matcher) bool {
	return m.MapIndex() == other.MapIndex()
}

func (m Matcher) Matches(o client.Object) bool {
	if m.Namespace != "" && o.GetNamespace() != m.Namespace {
		return false
	}
	if m.Name != "" && o.GetName() != m.Name {
		return false
	}
	// by default select is Empty which means match everything
	var sel metav1.LabelSelector
	if m.Selector != nil {
		sel = *m.Selector
	}
	selector, err := metav1.LabelSelectorAsSelector(&sel) // nil means select nothing
	if err != nil {
		return false
	}
	if !selector.Matches(labels.Set(o.GetLabels())) {
		return false
	}
	if m.Owner != nil {
		var found bool
		for _, owner := range o.GetOwnerReferences() {
			if owner.APIVersion == m.Owner.APIVersion &&
				owner.Kind == m.Owner.Kind &&
				owner.Name == m.Owner.Name &&
				matchesController(owner.Controller, m.Owner.Controller) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func matchesController(actual, expected *bool) bool {
	if expected != nil && *expected {
		return actual != nil && *actual
	}
	return false
}
