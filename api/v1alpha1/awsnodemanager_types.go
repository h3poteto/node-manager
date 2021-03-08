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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AWSNodeManagerSpec defines the desired state of AWSNodeManager
type AWSNodeManagerSpec struct {
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
	// +kubebuilder:validation:Type:=boolean
	// +kubebuilder:default=true
	EnableReplenish bool `json:"enableReplenish"`
	// +optional
	// +kubebuilder:validation:Type:=string
	RefreshSchedule string `json:"refreshSchedule"`
}

// AWSNodeManagerStatus defines the observed state of AWSNodeManager
type AWSNodeManagerStatus struct {
	// +nullable
	NodeReplenisher *AWSNodeReplenisherRef `json:"nodeReplenisher"`
	// +nullable
	NodeRefresher *AWSNodeRefresherRef `json:"nodeRefresher"`
	AWSNodes      []AWSNode            `json:"awsNodes,omitempty"`
	// +optinal
	// +nullable
	LastASGModifiedTime *metav1.Time `json:"lastASGModifiedTime,omitempty"`
	// +kubebuilder:default=0
	Revision int64 `json:"revision"`
	// +kubebuilder:default=init
	Phase AWSNodeManagerPhase `json:"phase"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// AWSNodeManager is the Schema for the awsnodemanagers API
type AWSNodeManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSNodeManagerSpec   `json:"spec,omitempty"`
	Status AWSNodeManagerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AWSNodeManagerList contains a list of AWSNodeManager
type AWSNodeManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSNodeManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSNodeManager{}, &AWSNodeManagerList{})
}

type AWSNode struct {
	Name                 string      `json:"name"`
	InstanceID           string      `json:"instanceID"`
	AvailabilityZone     string      `json:"availabilityZone"`
	InstanceType         string      `json:"instanceType"`
	AutoScalingGroupName string      `json:"autoScalingGroupName"`
	CreationTimestamp    metav1.Time `json:"creationTimestamp"`
}

type NodeRole string

const (
	Master = NodeRole("master")
	Worker = NodeRole("worker")
)

type AWSNodeManagerPhase string

const (
	AWSNodeManagerInit         = AWSNodeManagerPhase("init")
	AWSNodeManagerSynced       = AWSNodeManagerPhase("synced")
	AWSNodeManagerRefreshing   = AWSNodeManagerPhase("refreshing")
	AWSNodeManagerReplenishing = AWSNodeManagerPhase("replenishing")
)

type AWSNodeReplenisherRef struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	Namespace string `json:"namespace"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	Name string `json:"name"`
}

type AWSNodeRefresherRef struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	Namespace string `json:"namespace"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	Name string `json:"name"`
}
