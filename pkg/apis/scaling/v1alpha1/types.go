package v1alpha1

import (
	"time"

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
	CurrentReplicas int32     `json:"currentReplicas"`
	DesiredReplicas int32     `json:"desiredReplicas"`
	LastScaleTime   metav1.Time `json:"lastScaleTime"`
}

type SchedulingScalerSpec struct {
	Schedules      []SchedulingScalerSchedule  `json:"schedules"`
	ScaleTargetRef CrossVersionObjectReference `json:"scaleTargetRef"`
}

type SchedulingScalerSchedule struct {
	ScheduleTime SchedulingTime `json:"scheduleTime"`
	Replicas     int32          `json:"replicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SchedulingScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []SchedulingScaler `json:"items"`
}

type CrossVersionObjectReference struct {
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	APIVersion string `json:"apiVersion"`
}

type SchedulingTime struct {
	metav1.Time
}

func (st SchedulingTime) format() string {
	return st.Time.Format("15:04")
}

func (st SchedulingTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + st.format() + `"`), nil
}

func (st *SchedulingTime) UnmarshalJSON(data []byte) (err error) {
	t, err := time.Parse("15:04", string(data))
	if err != nil {
		return err
	}
	st = &SchedulingTime{metav1.Time{t}}
	return
}
