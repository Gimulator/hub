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

// MLSpec defines the desired state of ML
type MLSpec struct {
	RunID           int    `json:"run-id"`
	SubmissionID    int    `json:"submission-id"`
	EvaluatorImage  string `json:"evaluator-image"`
	SubmissionImage string `json:"submission-image"`
	BackoffLimit    int32  `json:"backoff-limit,omitempty"`

	CPUResourceRequest       string `json:"cpu-resource-request"`
	MemoryResourceRequest    string `json:"memory-resource-request"`
	EphemeralResourceRequest string `json:"ephemeral-resource-request"`
	CPUResourceLimit         string `json:"cpu-resource-limit"`
	MemoryResourceLimit      string `json:"memory-resource-limit"`
	EphemeralResourceLimit   string `json:"ephemeral-resource-limit"`

	DataPersistentVolumeClaimName       string `json:"data-persist-volume-claim-name"`
	EvaluationPersistentVolumeClaimName string `json:"evaluation-persist-volume-claim-name"`
}

// MLStatus defines the observed state of ML
type MLStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// ML is the Schema for the mls API
type ML struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MLSpec   `json:"spec,omitempty"`
	Status MLStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MLList contains a list of ML
type MLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ML `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ML{}, &MLList{})
}
