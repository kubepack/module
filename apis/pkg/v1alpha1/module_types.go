package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kmapi "kmodules.xyz/client-go/api/v1"
	rsapi "kmodules.xyz/resource-metadata/apis/meta/v1alpha1"
)

// ModuleSpec defines the desired state of Module
type ModuleSpec struct {
	Actions  []Action          `json:"actions"`
	EdgeList []rsapi.NamedEdge `json:"edgeList,omitempty"`
}

// Check array, map, etc
// can this be always string like in --set keys?
// Keep is such that we can always generate helm equivalent command
// https://helm.sh/docs/intro/using_helm/#the-format-and-limitations-of---set
type KV struct {
	Key      string   `json:"key"`
	ValueRef ValueRef `json:"value"`
}

type ValueRef struct {
	// string, nil, null
	Type string `json:"type"`
	// format is an optional OpenAPI type definition for this column. The 'name' format is applied
	// to the primary identifier column to assist in clients identifying column is the resource name.
	// See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#data-types for more.
	// +optional
	// Format string `json:"format,omitempty"`

	Raw string `json:"raw,omitempty"`

	// FieldPathTemplate is a Go text template that will be evaluated to determine cell value.
	// Users can use JSONPath expression to extract nested fields and apply template functions from Masterminds/sprig library.
	// The template function for JSON path is called `jp`.
	// Example: {{ jp "{.a.b}" . }} or {{ jp "{.a.b}" true }}, if json output is desired from JSONPath parser
	// +optional
	FieldPathTemplate string `json:"fieldPathTemplate,omitempty"`
	//
	//
	// Directly use path from object
	// +optional
	FieldPath string `json:"fieldPath,omitempty"`

	// json patch operation
	// See also: http://jsonpatch.com/
	// Op string `json:"op"`
}

type LoadValue struct {
	ObjRef *ObjectLocator `json:"objref,omitempty"`
	Values []KV           `json:"values"`
}

type ObjectLocator struct {
	Src   ObjectRef `json:"src"`
	Paths []string  `json:"paths,omitempty"` // sequence of DirectedEdge names
}

type ObjectRef struct {
	Target       metav1.TypeMeta       `json:"target"`
	Selector     *metav1.LabelSelector `json:"selector,omitempty"`
	Name         string                `json:"name,omitempty"`
	NameTemplate string                `json:"nameTemplate,omitempty"`
	UseAction    string                `json:"useAction"` // Use the values from that release == action to render templates
}

type Action struct {
	// Also the release name
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	ChartRef rsapi.ChartRepoRef `json:"chartRef,omitempty"`

	ValuesFile string `json:"valuesFile,omitempty"`
	// RFC 6902 compatible json patch. ref: http://jsonpatch.com
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	ValuesPatch *runtime.RawExtension `json:"valuesPatch,omitempty"`

	ValueOverrides []LoadValue `json:"overrideValues,omitempty"`

	// https://github.com/tamalsaha/kstatus-demo
	ReadinessCriteria ReadinessCriteria `json:"readinessCriteria"`

	Prerequisites *Prerequisites `json:"prerequisites,omitempty"`

	// ReportStatus []ObjectLocator
}

type Prerequisites struct {
	RequiredResources []metav1.GroupVersionResource `json:"requiredResources,omitempty"`
}

type ReadinessCriteria struct {
	Timeout metav1.Duration `json:"timeout"`

	// List objects for which to wait to reconcile using kstatus == Current
	// Same as helm --wait
	WaitForReconciled bool `json:"waitForReconciled"`

	ResourcesExist []metav1.GroupVersionResource `json:"requiredResources,omitempty"`
	WaitFors       []WaitFlags                   `json:"waitFors,omitempty"`
}

type ChartRef struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// wait ([-f FILENAME] | resource.group/resource.name | resource.group [(-l label | --all)]) [--for=delete|--for condition=available]

type WaitFlags struct {
	Resource     metav1.GroupResource  `json:"resource"`
	Labels       *metav1.LabelSelector `json:"labels"`
	All          bool                  `json:"all"`
	Timeout      metav1.Duration       `json:"timeout"`
	ForCondition string                `json:"for"`
}

// ModuleStatus defines the observed state of Module
type ModuleStatus struct {
	// Specifies the current phase of the database
	// +optional
	Phase string `json:"phase,omitempty"`
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions applied to the database, such as approval or denial.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Module is the Schema for the flows API
type Module struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModuleSpec   `json:"spec,omitempty"`
	Status ModuleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ModuleList contains a list of Module
type ModuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Module `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Module{}, &ModuleList{})
}
