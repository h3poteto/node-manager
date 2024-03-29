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

// NodeManagerSpec defines the desired state of NodeManager
type NodeManagerSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type:=string
	// +kubebuilder:default=aws
	// +kubebuilder:validation:Enum=aws
	CloudProvider string `json:"cloudProvider"`
	// +nullable
	Aws *CloudAWS `json:"aws,omitempty"`
}

// NodeManagerStatus defines the observed state of NodeManager
type NodeManagerStatus struct {
	// +nullable
	MasterAWSNodeManager *AWSNodeManagerRef `json:"masterAWSNodeManager,omitempty"`
	// +nullable
	WorkerAWSNodeManager *AWSNodeManagerRef `json:"workerAWSNodeManager,omitempty"`
	MasterNodes          []string           `json:"masterNodes,omitempty"`
	WorkerNodes          []string           `json:"workerNodes,omitempty"`
}

// +kubebuilder:object:root=true

// NodeManager is the Schema for the nodemanagers API
type NodeManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeManagerSpec   `json:"spec,omitempty"`
	Status NodeManagerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NodeManagerList contains a list of NodeManager
type NodeManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeManager{}, &NodeManagerList{})
}

type CloudAWS struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type:=string
	Region string `json:"region"`
	// +nullable
	Masters *Nodes `json:"masters,omitempty"`
	// +nullable
	Workers *Nodes `json:"workers,omitempty"`
}

type Nodes struct {
	// +kubebuilder:validation:Required
	AutoScalingGroups []AutoScalingGroup `json:"autoScalingGroups"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type:=integer
	Desired int32 `json:"desired"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type:=integer
	ASGModifyCoolTimeSeconds int64 `json:"asgModifyCoolTimeSeconds"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type:=boolean
	// +kubebuilder:default=true
	EnableReplenish bool `json:"enableReplenish"`
	// +nullable
	// +kubebuilder:validation:Type:=string
	RefreshSchedule string `json:"refreshSchedule"`
	// +optional
	// +kubebuilder:validation:Type:=integer
	// +kubebuilder:default=1
	SurplusNodes int64 `json:"surplusNodes"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=integer
	DrainGracePeriodSeconds int64 `json:"drainGracePeriodSeconds"`
}

type AutoScalingGroup struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type:=string
	Name string `json:"name"`
}

type AWSNodeManagerRef struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	Namespace string `json:"namespace"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	Name string `json:"name"`
}
