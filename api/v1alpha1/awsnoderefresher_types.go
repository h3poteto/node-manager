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

// AWSNodeRefresherSpec defines the desired state of AWSNodeRefresher
type AWSNodeRefresherSpec struct {
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
	ASGModifyCoolTimeSeconds int64 `json:"asgModfyCoolTimeSeconds"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type:=string
	Role NodeRole `json:"role"`
	// +kubebuilder:validation:Required
	// +kubebuilder:valitation:Type:=string
	Schedule string `json:"schedule"`
}

// AWSNodeRefresherStatus defines the observed state of AWSNodeRefresher
type AWSNodeRefresherStatus struct {
	AWSNodes []AWSNode `json:"awsNodes,omitempty"`
	// +optinal
	// +nullable
	LastASGModifiedTime *metav1.Time `json:"lastASGModifiedTime,omitempty"`
	// +kubebuilder:default=0
	Revision int64 `json:"revision"`
	// +kubebuilder:default=init
	Phase AWSNodeRefresherPhase `json:"phase"`
	// +optinal
	// +nullable
	NextUpdateTime *metav1.Time `json:"nextUpdateTime"`
	// +optinal
	// +nullable
	UpdateStartTime *metav1.Time `json:"updateStartTime"`
	// +optional
	// +nullable
	ReplaceTargetNode *AWSNode `json:"replaceTargetNode,omitempty"`
}

// +kubebuilder:object:root=true

// AWSNodeRefresher is the Schema for the awsnoderefreshers API
type AWSNodeRefresher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSNodeRefresherSpec   `json:"spec,omitempty"`
	Status AWSNodeRefresherStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AWSNodeRefresherList contains a list of AWSNodeRefresher
type AWSNodeRefresherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSNodeRefresher `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSNodeRefresher{}, &AWSNodeRefresherList{})
}

type AWSNodeRefresherPhase string

const (
	AWSNodeRefresherInit             = AWSNodeRefresherPhase("init")
	AWSNodeRefresherScheduled        = AWSNodeRefresherPhase("scheduled")
	AWSNodeRefresherUpdateIncreasing = AWSNodeRefresherPhase("increasing")
	AWSNodeRefresherUpdateReplacing  = AWSNodeRefresherPhase("replacing")
	AWSNodeRefresherUpdateAWSWaiting = AWSNodeRefresherPhase("awsWaiting")
	AWSNodeRefresherUpdateDecreasing = AWSNodeRefresherPhase("decreasing")
	AWSNodeRefresherCompleted        = AWSNodeRefresherPhase("completed")
)
