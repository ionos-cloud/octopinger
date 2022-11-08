package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// CRDResourceKind ...
	CRDResourceKind = "Octopinger"
)

func init() {
	SchemeBuilder.Register(&Octopinger{}, &OctopingerList{})
}

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make generate" or "go generate ./..." to regenerate code after modifying this file

//+kubebuilder:object:root=true

// Octopinger is the Schema for the octopinger API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +operator-sdk:csv:customresourcedefinitions:resources={{Octopinger,v1alpha1,""}}
// +operator-sdk:csv:customresourcedefinitions:resources={{Pod,v1,""}}
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
	// Label is the value of the 'octopinger=' label to set on a node that should run Octopinger.
	Label string `json:"label"`

	// Config is a wrapper to contain the configuration for Octopinger.
	Config Config `json:"config"`

	// Template specifies the options for the DaemonSet template.
	Template Template `json:"template"`
}

// Config is a wrapper to contain the configuration of Octopinger.
type Config struct {
	// ICMP is the configuration for the ICMP probe.
	ICMP ICMP `json:"icmp"`

	// DNS is the configuration for the DNS probe.
	DNS DNS `json:"dns"`
}

// DNS configures this probe.
type DNS struct {
	// Enable is turning the DNS probe of for Octopinger.
	Enable bool `json:"enable"`
	// Names contains the list of domain names to query.
	Names []string `json:"names,omitempty"`
	// Server contains a domain name servers to use for the probe. By default the configured DNS servers are used.
	Server string `json:"server,omitempty"`
	// Timeout the time to wait for the probe to succeed. The default is "1m" (1 minute).
	Timeout string `json:"timeout,omitempty"`
}

// ICMP configures this probe.
type ICMP struct {
	// Enable is turning the ICMP probe on for Octopinger. By default all nodes are probed.
	Enable bool `json:"enable"`
	// ExcludeNodes allows to exclude specific nodes from probing.
	ExcludeNodes []string `json:"exclude_nodes,omitempty"`
	// AdditionalTargets this is a list of additional targets to probe via ICMP.
	AdditionalTargets []string ` json:"additionaltargets,omitempty"`
	// Timeout the time to wait for the probe to succeed. The default is "1m" (1 minute).
	Timeout string `json:"timeout,omitempty"`
	// Count is number of ICMP packets to send.
	Count int `json:"count,omitempty"`
	// NodePacketLossThreshold determines the threshold to report a node as available or not (Default: "0.05")
	NodePacketLossThreshold string `json:"node_packet_loss_treshold,omitempty"`
}

// Template ...
type Template struct {
	// Image is the Docker image to run for octopinger.
	Image string `json:"image"`

	// Tolerations ...
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

//+kubebuilder:object:root=true

// OctopingerList contains a list of Octopinger
type OctopingerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Octopinger `json:"items"`
}

type OctopingerPhase string

const (
	OctopingerPhaseNone     OctopingerPhase = ""
	OctopingerPhaseCreating OctopingerPhase = "Creating"
	OctopingerPhaseRunning  OctopingerPhase = "Running"
	OctopingerPhaseFailed   OctopingerPhase = "Failed"
)

// OctopingerStatus defines the observed state of Octopinger
// +k8s:openapi-gen=true
type OctopingerStatus struct {
	// Phase is the octopinger running phase.
	Phase OctopingerPhase `json:"phase"`

	// ControlPaused indicates the operator pauses the control of
	// Octopinger.
	ControlPaused bool `json:"controlPaused,omitempty"`
}

// IsFailed ...
func (cs *OctopingerStatus) IsFailed() bool {
	if cs == nil {
		return false
	}

	return cs.Phase == OctopingerPhaseFailed
}

// SetPhase ...
func (cs *OctopingerStatus) SetPhase(p OctopingerPhase) {
	cs.Phase = p
}

// PauseControl ...
func (cs *OctopingerStatus) PauseControl() {
	cs.ControlPaused = true
}

// Control ...
func (cs *OctopingerStatus) Control() {
	cs.ControlPaused = false
}
