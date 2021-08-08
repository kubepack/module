package api

import (
	"encoding/gob"
	"github.com/cespare/xxhash"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +genset=true
type NamespacedName struct {
	Namespace string
	Name      string
}

// +genset=true
type Matcher struct {
	Name      string
	Namespace string
	Selector  *metav1.LabelSelector
}

func (m *Matcher) MapIndex() uint64 {
	if m == nil {
		return 0
	}
	h := xxhash.New()
	_ = gob.NewEncoder(h).Encode(m)
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
	return selector.Matches(labels.Set(o.GetLabels()))
}
