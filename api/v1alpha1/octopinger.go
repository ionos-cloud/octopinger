package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&Octopinger{}, &OctopingerList{})
}

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make" to regenerate code after modifying this file

//+kubebuilder:object:root=true

// Octopinger is the Schema for the octopinger API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +operator-sdk:csv:customresourcedefinitions:resources={{Octopinger,v1alpha1,""}}
// +operator-sdk:csv:customresourcedefinitions:resources={{Pod,v1,""}}
// +operator-sdk:csv:customresourcedefinitions:resources={{Prometheus,v1,""}}
// +operator-sdk:csv:customresourcedefinitions:resources={{ReplicaSet,v1,""}}
type Octopinger struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OctopingerSpec   `json:"spec,omitempty"`
	Status OctopingerStatus `json:"status,omitempty"`
}

// OctopingerSpec defines the desired state of Octopinger
// +k8s:openapi-gen=true
type OctopingerSpec struct {
}

//+kubebuilder:object:root=true

// OctopingerList contains a list of Octopinger
type OctopingerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Octopinger `json:"items"`
}

// OctopingerStatus defines the observed state of Octopinger
// +k8s:openapi-gen=true
type OctopingerStatus struct{}
