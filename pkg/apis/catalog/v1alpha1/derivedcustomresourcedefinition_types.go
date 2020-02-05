/*
Copyright 2019 The KubeCarrier Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// DerivedCustomResourceDefinitionSpec defines the desired state of DerivedCustomResourceDefinition.
type DerivedCustomResourceDefinitionSpec struct {
	// CRD that should be used as a base to derive a new CRD from.
	BaseCRD ObjectReference `json:"baseCRD"`
	// overrides the kind of the derived CRD.
	KindOverride string `json:"kindOverride,omitempty"`
	// controls which fields will be present in the derived CRD.
	Expose []VersionExposeConfig `json:"expose"`
}

// VersionExposeConfig specifies which fields to expose in the derived CRD.
type VersionExposeConfig struct {
	// specifies the versions of the referenced CRD, that this expose config applies to.
	// The same version may not be specified in multiple VersionExposeConfigs.
	Versions []string `json:"versions"`
	// specifies the fields that should be present in the derived CRD.
	Fields []FieldPath `json:"fields"`
}

// FieldPath is specifying how to address a certain field.
type FieldPath struct {
	// JSONPath e.g. .spec.somefield.somesubfield
	JSONPath string `json:"jsonPath"`
}

// DerivedCustomResourceDefinitionStatus defines the observed state of DerivedCustomResourceDefinition.
type DerivedCustomResourceDefinitionStatus struct {
	// ObservedGeneration is the most recent generation observed for this DerivedCustomResourceDefinition by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a DerivedCustomResourceDefinition's current state.
	Conditions []DerivedCustomResourceDefinitionCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase DerivedCustomResourceDefinitionPhaseType `json:"phase,omitempty"`
	// DerivedCRD holds information about the derived CRD.
	DerivedCRD *DerivedCustomResourceDefinitionReference `json:"derivedCRD,omitempty"`
}

// DerivedCustomResourceDefinitionReference references the derived CRD controlled by this DerivedCustomResourceDefinition instance.
type DerivedCustomResourceDefinitionReference struct {
	// Name of the derived CRD.
	Name string `json:"name"`
	// API Group of the derived CRD.
	Group    string `json:"group"`
	Kind     string `json:"kind"`
	Plural   string `json:"plural"`
	Singular string `json:"singular"`
}

// DerivedCustomResourceDefinitionPhaseType represents all conditions as a single string for printing by using kubectl commands.
type DerivedCustomResourceDefinitionPhaseType string

// Values of DerivedCustomResourceDefinitionPhaseType.
const (
	DerivedCustomResourceDefinitionPhaseReady    DerivedCustomResourceDefinitionPhaseType = "Ready"
	DerivedCustomResourceDefinitionPhaseNotReady DerivedCustomResourceDefinitionPhaseType = "NotReady"
	DerivedCustomResourceDefinitionPhaseUnknown  DerivedCustomResourceDefinitionPhaseType = "Unknown"
)

// updatePhase updates the phase property based on the current conditions
// this method should be called every time the conditions are updated.
func (s *DerivedCustomResourceDefinitionStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != DerivedCustomResourceDefinitionReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = DerivedCustomResourceDefinitionPhaseReady
		case ConditionFalse:
			s.Phase = DerivedCustomResourceDefinitionPhaseNotReady
		case ConditionUnknown:
			s.Phase = DerivedCustomResourceDefinitionPhaseUnknown
		}
		return
	}

	s.Phase = DerivedCustomResourceDefinitionPhaseUnknown
}

// DerivedCustomResourceDefinitionConditionType represents a DerivedCustomResourceDefinitionCondition value.
type DerivedCustomResourceDefinitionConditionType string

const (
	// DerivedCustomResourceDefinitionReady represents a DerivedCustomResourceDefinition condition is in ready state.
	DerivedCustomResourceDefinitionReady DerivedCustomResourceDefinitionConditionType = "Ready"
	// DerivedCustomResourceDefinitionEstablished is True if the derived crd could be registered and is now served by the kube-apiserver.
	DerivedCustomResourceDefinitionEstablished DerivedCustomResourceDefinitionConditionType = "Established"
	// DerivedCustomResourceDefinitionControllerReady is Ture if the controller to propagate the derived and internal crd is ready.
	DerivedCustomResourceDefinitionControllerReady DerivedCustomResourceDefinitionConditionType = "ControllerReady"
)

// DerivedCustomResourceDefinitionCondition contains details for the current condition of this DerivedCustomResourceDefinition.
type DerivedCustomResourceDefinitionCondition struct {
	// Type is the type of the DerivedCustomResourceDefinition condition, currently ('Ready').
	Type DerivedCustomResourceDefinitionConditionType `json:"type"`
	// Status is the status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// LastTransitionTime is the last time the condition transits from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *DerivedCustomResourceDefinitionStatus) GetCondition(t DerivedCustomResourceDefinitionConditionType) (condition DerivedCustomResourceDefinitionCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// SetCondition replaces or adds the given condition.
func (s *DerivedCustomResourceDefinitionStatus) SetCondition(condition DerivedCustomResourceDefinitionCondition) {
	defer s.updatePhase()

	if condition.LastTransitionTime.IsZero() {
		condition.LastTransitionTime = metav1.Now()
	}

	for i := range s.Conditions {
		if s.Conditions[i].Type == condition.Type {

			// Only update the LastTransitionTime when the Status is changed.
			if s.Conditions[i].Status != condition.Status {
				s.Conditions[i].LastTransitionTime = condition.LastTransitionTime
			}

			s.Conditions[i].Status = condition.Status
			s.Conditions[i].Reason = condition.Reason
			s.Conditions[i].Message = condition.Message

			return
		}
	}

	s.Conditions = append(s.Conditions, condition)
}

// DerivedCustomResourceDefinition derives a new CRD from a existing one.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=dcrd
type DerivedCustomResourceDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DerivedCustomResourceDefinitionSpec   `json:"spec,omitempty"`
	Status DerivedCustomResourceDefinitionStatus `json:"status,omitempty"`
}

// DerivedCustomResourceDefinitionList contains a list of DerivedCustomResourceDefinition.
// +kubebuilder:object:root=true
type DerivedCustomResourceDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DerivedCustomResourceDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DerivedCustomResourceDefinition{}, &DerivedCustomResourceDefinitionList{})
}
