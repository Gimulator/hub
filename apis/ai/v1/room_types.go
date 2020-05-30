package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&Room{}, &RoomList{})
}

// RoomSpec defines the desired state of Room
type RoomSpec struct {
	ID                    int         `json:"id"`
	BackoffLimit          int32       `json:"backoff-limit,omitempty"`
	ActiveDeadLineSeconds int64       `json:"active-dead-line-seconds,omitempty"`
	Sketch                string      `json:"sketch"`
	Actors                []Actor     `json:"actors"`
	Volumes               []Volume    `json:"volumes,omitempty"`
	ConfigMaps            []ConfigMap `json:"config-maps,omitempty"`
}

type RoomStatusType string

const (
	RoomStatusTypeUnknown RoomStatusType = "unknown"
	RoomStatusTypeRunning RoomStatusType = "running"
	RoomStatusTypeFailed  RoomStatusType = "failed"
	RoomStatusTypeSuccess RoomStatusType = "success"
)

type NamespacedName struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// RoomStatus defines the observed state of Room
type RoomStatus struct {
	ConfigMapList []NamespacedName `json:"configmap-list"`
	JobList       []NamespacedName `json:"job-list"`

	RoomStatusType RoomStatusType `json:"room-status-type"`
}

func (r *RoomStatus) AddJob(new NamespacedName) {
	if r.JobList == nil {
		r.JobList = make([]NamespacedName, 0)
	}

	for _, nn := range r.JobList {
		if nn.Name == new.Name && nn.Namespace == new.Namespace {
			return
		}
	}
	r.JobList = append(r.JobList, new)
}

func (r *RoomStatus) AddConfigMap(new NamespacedName) {
	if r.ConfigMapList == nil {
		r.ConfigMapList = make([]NamespacedName, 0)
	}

	for _, nn := range r.ConfigMapList {
		if nn.Name == new.Name && nn.Namespace == new.Namespace {
			return
		}
	}
	r.ConfigMapList = append(r.ConfigMapList, new)
}

// +kubebuilder:object:root=true

// Room is the Schema for the rooms API
type Room struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoomSpec   `json:"spec,omitempty"`
	Status RoomStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RoomList contains a list of Room
type RoomList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Room `json:"items"`
}
