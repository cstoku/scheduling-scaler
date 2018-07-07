package v1alpha

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SchedulingScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status SchedulingScalerStatus `json:"status"`
	Spec   SchedulingScalerSpec   `json:"spec"`
}

type SchedulingScalerStatus struct {
	Name string `json:"name"`
}

type SchedulingScalerSpec struct {
	Name string `json:"name"`
}
