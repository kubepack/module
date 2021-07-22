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
	EdgeList []rsapi.NamedEdge `json:"edge_list"`
}

// Check array, map, etc
// can this be always string like in --set keys?
// Keep is such that we can always generate helm equivalent command
type KV struct {
	Key string `json:"key"`
	// string, nil, null
	Type string `json:"type"`
	// format is an optional OpenAPI type definition for this column. The 'name' format is applied
	// to the primary identifier column to assist in clients identifying column is the resource name.
	// See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#data-types for more.
	// +optional
	// Format string `json:"format,omitempty"`

	// PathTemplate is a Go text template that will be evaluated to determine cell value.
	// Users can use JSONPath expression to extract nested fields and apply template functions from Masterminds/sprig library.
	// The template function for JSON path is called `jp`.
	// Example: {{ jp "{.a.b}" . }} or {{ jp "{.a.b}" true }}, if json output is desired from JSONPath parser
	// +optional
	PathTemplate string `json:"pathTemplate,omitempty"`
	//
	//
	// Directly use path from object
	// +optional
	Path string `json:"path,omitempty"`

	// json patch operation
	// See also: http://jsonpatch.com/
	// Op string `json:"op"`
}

type LoadValue struct {
	From   ObjectLocator `json:"from"`
	Values []KV          `json:"values"`
}

type ObjectLocator struct {
	// Use the values from that release == action to render templates
	UseRelease string    `json:"use_release"`
	Src        ObjectRef `json:"src"`
	Paths      []string  `json:"paths"` // sequence of DirectedEdge names
}

type ObjectRef struct {
	Target       metav1.TypeMeta       `json:"target"`
	Selector     *metav1.LabelSelector `json:"selector,omitempty"`
	Name         string                `json:"name,omitempty"`
	NameTemplate string                `json:"nameTemplate,omitempty"`
}

type Action struct {
	// Also the action name
	ReleaseName string `json:"releaseName" protobuf:"bytes,3,opt,name=releaseName"`

	rsapi.ChartRepoRef `json:",inline" protobuf:"bytes,1,opt,name=chartRef"`

	ValuesFile string `json:"valuesFile,omitempty" protobuf:"bytes,6,opt,name=valuesFile"`
	// RFC 6902 compatible json patch. ref: http://jsonpatch.com
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	ValuesPatch *runtime.RawExtension `json:"valuesPatch,omitempty" protobuf:"bytes,7,opt,name=valuesPatch"`

	ValueOverrides []LoadValue `json:"overrideValues"`

	// https://github.com/tamalsaha/kstatus-demo
	ReadinessCriteria ReadinessCriteria `json:"readiness_criteria"`

	Prerequisites Prerequisites `json:"prerequisites"`
}

type Prerequisites struct {
	RequiredResources []metav1.GroupVersionResource `json:"required_resources"`
}

type ReadinessCriteria struct {
	Timeout metav1.Duration `json:"timeout"`

	// List objects for which to wait to reconcile using kstatus == Current
	// Same as helm --wait
	WaitForReconciled bool `json:"wait_for_reconciled"`

	ResourcesExist []metav1.GroupVersionResource `json:"required_resources"`
	WaitFors       []WaitFlags                   `json:"waitFors,omitempty" protobuf:"bytes,9,rep,name=waitFors"`
}

type ChartRef struct {
	URL  string `json:"url" protobuf:"bytes,1,opt,name=url"`
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

// wait ([-f FILENAME] | resource.group/resource.name | resource.group [(-l label | --all)]) [--for=delete|--for condition=available]

type WaitFlags struct {
	Resource     metav1.GroupResource  `json:"resource" protobuf:"bytes,1,opt,name=resource"`
	Labels       *metav1.LabelSelector `json:"labels" protobuf:"bytes,2,opt,name=labels"`
	All          bool                  `json:"all" protobuf:"varint,3,opt,name=all"`
	Timeout      metav1.Duration       `json:"timeout" protobuf:"bytes,4,opt,name=timeout"`
	ForCondition string                `json:"for" protobuf:"bytes,5,opt,name=for"`
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
