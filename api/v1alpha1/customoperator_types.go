package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BGDeploymentOperatorSpec defines the desired state of BGDeploymentOperator
type BGDeploymentOperatorSpec struct {
	Replicas int32 `json:"replicas"`
	Image string `json:"image"`
}

// BGDeploymentOperatorStatus defines the observed state of BGDeploymentOperator
type BGDeploymentOperatorStatus struct {
	ActiveColor string `json:"activeColor"`
}

// BGDeploymentOperator is the Schema for the bgdeploymentoperators API
type BGDeploymentOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BGDeploymentOperatorSpec   `json:"spec,omitempty"`
	Status BGDeploymentOperatorStatus `json:"status,omitempty"`
}

// BGDeploymentOperatorList contains a list of BGDeploymentOperator
type BGDeploymentOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BGDeploymentOperator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BGDeploymentOperator{}, &BGDeploymentOperatorList{})
}
