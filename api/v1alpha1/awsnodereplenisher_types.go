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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AWSNodeReplenisherSpec defines the desired state of AWSNodeReplenisher
type AWSNodeReplenisherSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type:=string
	Region string `json:"region"`
	// +kubebuilder:validation:Required
	AutoScalingGroups []AutoScalingGroup `json:"autoScalingGroups"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type:=integer
	Desired int32 `json:"desired"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type:=integer
	ASGModifyCoolTimeSeconds int64 `json:"asgModifyCoolTimeSeconds"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type:=string
	Role NodeRole `json:"role"`
}

// AWSNodeReplenisherStatus defines the observed state of AWSNodeReplenisher
type AWSNodeReplenisherStatus struct {
	AWSNodes          []AWSNode `json:"awsNodes,omitempty"`
	NotJoinedAWSNodes []AWSNode `json:"notJoinedAWSNodes,omitempty"`
	// +optinal
	// +nullable
	LastASGModifiedTime *metav1.Time `json:"lastASGModifiedTime,omitempty"`
	// +kubebuilder:default=0
	Revision int64 `json:"revision"`
	// +kubebuilder:default=init
	Phase AWSNodeReplenisherPhase `json:"phase"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// AWSNodeReplenisher is the Schema for the awsnodereplenishers API
type AWSNodeReplenisher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSNodeReplenisherSpec   `json:"spec,omitempty"`
	Status AWSNodeReplenisherStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AWSNodeReplenisherList contains a list of AWSNodeReplenisher
type AWSNodeReplenisherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSNodeReplenisher `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSNodeReplenisher{}, &AWSNodeReplenisherList{})
}

type AWSNodeReplenisherPhase string

const (
	AWSNodeReplenisherInit        = AWSNodeReplenisherPhase("init")
	AWSNodeReplenisherSynced      = AWSNodeReplenisherPhase("synced")
	AWSNodeReplenisherAWSUpdating = AWSNodeReplenisherPhase("awsUpdating")
)
