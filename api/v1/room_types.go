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

type PVCNames struct {
	Public  []string `json:"public,omitempty" yaml:"public,omitempty"`
	Private []string `json:"private,omitempty" yaml:"private,omitempty"`
}

type GimulatorSettings struct {
	Image     string                       `json:"image" yaml:"image"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty" yaml:"resources,omitempty"`
}

type RoleSettings struct {
	Resources *corev1.ResourceRequirements `json:"resources,omitempty" yaml:"resources,omitempty"`
}

type Setting struct {
	DataPVCNames     *PVCNames                   `json:"dataPVCNames,omitempty" yaml:"dataPVCNames,omitempty"`
	Gimulator        *GimulatorSettings          `json:"gimulator" yaml:"gimulator"`
	OutputVolumeSize string                      `json:"outputVolumeSize" yaml:"outputVolumeSize"`
	DefaultResources corev1.ResourceRequirements `json:"defaultResources" yaml:"defaultResources"`
	Roles            map[string]*RoleSettings    `json:"roles,omitempty" yaml:"roles,omitempty"`
	StorageClass     string                      `json:"storageClass" yaml:"storageClass"`
}

// Actor defines some actor of a Room
type Actor struct {
	Name      string                       `json:"name"`
	Image     string                       `json:"image"`
	Role      string                       `json:"role"`
	Token     string                       `json:"token,omitempty"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	Envs      []corev1.EnvVar              `json:"envs,omitempty"`
}

// Director defines the director of a Room
type Director struct {
	Name      string                       `json:"name"`
	Image     string                       `json:"image"`
	Token     string                       `json:"token,omitempty"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	Envs      []corev1.EnvVar              `json:"envs,omitempty"`
}

// RoomSpec defines the desired state of Room
type RoomSpec struct {
	ID                      string             `json:"id"`
	ProblemID               string             `json:"problemID"`
	Setting                 *Setting           `json:"setting,omitempty"`
	Gimulator               *GimulatorSettings `json:"gimulator,omitempty"`
	Actors                  []*Actor           `json:"actors"`
	Director                *Director          `json:"director"`
	Timeout                 uint64             `json:"timeout"`
	TerminateOnActorFailure bool               `json:"terminateOnActorFailure"`
}

// RoomStatus defines the observed state of Room
type RoomStatus struct {
	GimulatorStatus corev1.PodPhase            `json:"gimulatorStatus"`
	DirectorStatus  corev1.PodPhase            `json:"directorStatus"`
	ActorStatuses   map[string]corev1.PodPhase `json:"actorStatuses"`
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
