package v1alpha1

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
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
	// Label is the value of the 'octopinger=' label to set on a node that should run octopinger.
	Label string `json:"label"`

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
	// Types specifies the type of domain name records to query. By default this is A records.
	Types []string `json:"types,omitempty"`
	// Servers contains a list of domain name servers to use for the probe. By default the configured DNS servers are used.
	Servers []string `json:"servers,omitempty"`
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
	AdditionalTargets []string ` json:"additional_targets,omitempty"`
	// Timeout the time to wait for the probe to succeed. The default is "1m" (1 minute).
	Timeout string `json:"timeout,omitempty"`
	// TTL is the time to live for the ICMP packet.
	TTL string `json:"ttl,omitempty"`
}

// Template ...
type Template struct {
	// Image is the Docker image to run for octopinger.
	Image string `json:"image"`

	// Tolerations ...
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// Probes ...
type Probes []Probe

// ConfigMap ...
func (p Probes) ConfigMap() map[string]string {
	cfg := make(map[string]string)

	for _, probe := range p {
		maps.Copy(cfg, probe.ConfigMap())
	}

	delete(cfg, "")

	return cfg
}

// ConfigMap ...
func (p Probe) ConfigMap() map[string]string {
	cfg := map[string]string{
		fmt.Sprintf("probes.%s.enabled", p.Type):    strconv.FormatBool(p.Enabled),
		fmt.Sprintf("probes.%s.properties", p.Type): p.Properties.KeyValues(),
	}

	return cfg
}

// Properties ...
type Properties map[string]string

// KeyValues ...
func (p Properties) KeyValues() string {
	var lines []string
	for k, v := range p {
		lines = append(lines, fmt.Sprintf("%s:%v", k, v))
	}

	return strings.Join(lines, "\n")
}

// Probe ...
type Probe struct {
	// Name ...
	Name string `json:"name"`

	// Type ...
	Type string `json:"type"`

	// Enabled ...
	Enabled bool `json:"enabled"`

	// Properties ...
	Properties Properties `json:"properties,omitempty"`
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
