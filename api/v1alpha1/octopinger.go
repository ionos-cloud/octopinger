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
	// Label is the value of the 'octopinger=' label to set on a node that should run octopinger.
	Label string `json:"label"`

	// Image is the Docker image to run for octopinger.
	Image string `json:"image"`

	// Probes ...
	Probes Probes `json:"probes"`

	// Tolerations ...
	Tolerations []corev1.Toleration `json:"tolerations"`
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
	// Type ...
	Type string `json:"type"`

	// Enabled ...
	Enabled bool `json:"enabled"`

	// Properties ...
	Properties Properties `json:"properties"`
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

type OctopingerConditionType string

const (
	OctopingerConditionReady = "Ready"

	OctopingerConditionScalingUp   = "ScalingUp"
	OctopingerConditionScalingDown = "ScalingDown"

	OctopingerConditionUpgrading = "Upgrading"
)

type OctopingerCondition struct {
	Type OctopingerConditionType `json:"type"`

	Reason string `json:"reason"`

	TransitionTime string `json:"transitionTime"`
}

// OctopingerStatus defines the observed state of Octopinger
// +k8s:openapi-gen=true
type OctopingerStatus struct {
	// Phase is the octopinger running phase.
	Phase  OctopingerPhase `json:"phase"`
	Reason string          `json:"reason"`

	// ControlPaused indicates the operator pauses the control of
	// octopinger.
	ControlPaused bool `json:"controlPaused"`

	// Condition keeps ten most recent octopinger conditions.
	Conditions []OctopingerCondition `json:"conditions"`

	// Size is the number of nodes the daemon is deployed to.
	Size int `json:"size"`

	// CurrentVersion is the current octopinger version.
	CurrentVersion string `json:"currentVersion"`
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

// SetSize sets the current size of the cluster.
func (cs *OctopingerStatus) SetSize(size int) {
	cs.Size = size
}

// SetCurrentVersion ...
func (cs *OctopingerStatus) SetCurrentVersion(v string) {
	cs.CurrentVersion = v
}

// SetReason ...
func (cs *OctopingerStatus) SetReason(r string) {
	cs.Reason = r
}
