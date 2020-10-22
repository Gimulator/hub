/*
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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProblemSettings struct {
	DataPVCName            string `json:"dataPVCName,omitempty"`
	FactPVCName            string `json:"factPVCName,omitempty"`
	GimulatorImage         string `json:"gimulatorImage,omitempty"`
	OutputVolumeSize       string `json:"outputVolumeSize"`
	ResourceCPULimit       string `json:"resourceCPULimit"`
	ResourceMemoryLimit    string `json:"resourceMemoryLimit"`
	ResourceEphemeralLimit string `json:"resourceEphemeralLimit"`
}

// Actor defines some actor of a Room
type Actor struct {
	ID    string `json:"id"`
	Image string `json:"image"`
	Role  string `json:"role"`
	Token string `json:"token,omitempty"`
}

// Director defines the director of a Room
type Director struct {
	ID    string `json:"id"`
	Image string `json:"image"`
	Token string `json:"token,omitempty"`
}

// RoomSpec defines the desired state of Room
type RoomSpec struct {
	ID              string           `json:"id"`
	ProblemID       string           `json:"problemID"`
	ProblemSettings *ProblemSettings `json:"problemSettings,omitempty"`
	Actors          []Actor          `json:"actors"`
	Director        Director         `json:"director"`
}

// RoomStatus defines the observed state of Room
type RoomStatus struct {
	GimulatorStatus *corev1.PodStatus            `json:"gimulatorStatus"`
	DirectorStatus  *corev1.PodStatus            `json:"directorStatus"`
	ActorStatuses   map[string]*corev1.PodStatus `json:"actorStatuses"`
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

func init() {
	SchemeBuilder.Register(&Room{}, &RoomList{})
}
