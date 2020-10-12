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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Actor defines some actor of a Room
type Actor struct {
	ID    string `json:"id"`
	Image string `json:"image"`
}

// ActorStatus defines the observed state of Actor
type ActorStatus struct {
}

// Director defines the director of a Room
type Director struct {
	ID    string `json:"id"`
	Image string `json:"image"`
}

// DirectorStatus defines the observed state of Director
type DirectorStatus struct {
}

// RoomSpec defines the desired state of Room
type RoomSpec struct {
	ID        string   `json:"id"`
	Actors    []Actor  `json:"actors"`
	Director  Director `json:"director"`
	Gimulator bool     `json:"gimulator"`
	Volume    bool     `json:"volume"`
}

// RoomStatus defines the observed state of Room
type RoomStatus struct {
	DirectorStatus DirectorStatus `json:"directorStatus"`
	ActorStatuses  []ActorStatus  `json:"actorStatuses"`
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
