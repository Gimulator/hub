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
	RoomStatusTypeRunning RoomStatusType = "room-status-type-running"
	RoomStatusTypeFailed  RoomStatusType = "room-status-type-failed"
	RoomStatusTypeSuccess RoomStatusType = "room-status-type-success"
)

// RoomStatus defines the observed state of Room
type RoomStatus struct {
	RoomStatusType RoomStatusType `json:"room-status-type"`
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
