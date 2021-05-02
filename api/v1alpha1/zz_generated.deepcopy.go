// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNode) DeepCopyInto(out *AWSNode) {
	*out = *in
	in.CreationTimestamp.DeepCopyInto(&out.CreationTimestamp)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNode.
func (in *AWSNode) DeepCopy() *AWSNode {
	if in == nil {
		return nil
	}
	out := new(AWSNode)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNodeManager) DeepCopyInto(out *AWSNodeManager) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNodeManager.
func (in *AWSNodeManager) DeepCopy() *AWSNodeManager {
	if in == nil {
		return nil
	}
	out := new(AWSNodeManager)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSNodeManager) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNodeManagerList) DeepCopyInto(out *AWSNodeManagerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSNodeManager, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNodeManagerList.
func (in *AWSNodeManagerList) DeepCopy() *AWSNodeManagerList {
	if in == nil {
		return nil
	}
	out := new(AWSNodeManagerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSNodeManagerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNodeManagerRef) DeepCopyInto(out *AWSNodeManagerRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNodeManagerRef.
func (in *AWSNodeManagerRef) DeepCopy() *AWSNodeManagerRef {
	if in == nil {
		return nil
	}
	out := new(AWSNodeManagerRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNodeManagerSpec) DeepCopyInto(out *AWSNodeManagerSpec) {
	*out = *in
	if in.AutoScalingGroups != nil {
		in, out := &in.AutoScalingGroups, &out.AutoScalingGroups
		*out = make([]AutoScalingGroup, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNodeManagerSpec.
func (in *AWSNodeManagerSpec) DeepCopy() *AWSNodeManagerSpec {
	if in == nil {
		return nil
	}
	out := new(AWSNodeManagerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNodeManagerStatus) DeepCopyInto(out *AWSNodeManagerStatus) {
	*out = *in
	if in.NodeReplenisher != nil {
		in, out := &in.NodeReplenisher, &out.NodeReplenisher
		*out = new(AWSNodeReplenisherRef)
		**out = **in
	}
	if in.NodeRefresher != nil {
		in, out := &in.NodeRefresher, &out.NodeRefresher
		*out = new(AWSNodeRefresherRef)
		**out = **in
	}
	if in.AWSNodes != nil {
		in, out := &in.AWSNodes, &out.AWSNodes
		*out = make([]AWSNode, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.NotJoinedAWSNodes != nil {
		in, out := &in.NotJoinedAWSNodes, &out.NotJoinedAWSNodes
		*out = make([]AWSNode, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.LastASGModifiedTime != nil {
		in, out := &in.LastASGModifiedTime, &out.LastASGModifiedTime
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNodeManagerStatus.
func (in *AWSNodeManagerStatus) DeepCopy() *AWSNodeManagerStatus {
	if in == nil {
		return nil
	}
	out := new(AWSNodeManagerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNodeRefresher) DeepCopyInto(out *AWSNodeRefresher) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNodeRefresher.
func (in *AWSNodeRefresher) DeepCopy() *AWSNodeRefresher {
	if in == nil {
		return nil
	}
	out := new(AWSNodeRefresher)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSNodeRefresher) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNodeRefresherList) DeepCopyInto(out *AWSNodeRefresherList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSNodeRefresher, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNodeRefresherList.
func (in *AWSNodeRefresherList) DeepCopy() *AWSNodeRefresherList {
	if in == nil {
		return nil
	}
	out := new(AWSNodeRefresherList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSNodeRefresherList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNodeRefresherRef) DeepCopyInto(out *AWSNodeRefresherRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNodeRefresherRef.
func (in *AWSNodeRefresherRef) DeepCopy() *AWSNodeRefresherRef {
	if in == nil {
		return nil
	}
	out := new(AWSNodeRefresherRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNodeRefresherSpec) DeepCopyInto(out *AWSNodeRefresherSpec) {
	*out = *in
	if in.AutoScalingGroups != nil {
		in, out := &in.AutoScalingGroups, &out.AutoScalingGroups
		*out = make([]AutoScalingGroup, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNodeRefresherSpec.
func (in *AWSNodeRefresherSpec) DeepCopy() *AWSNodeRefresherSpec {
	if in == nil {
		return nil
	}
	out := new(AWSNodeRefresherSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNodeRefresherStatus) DeepCopyInto(out *AWSNodeRefresherStatus) {
	*out = *in
	if in.AWSNodes != nil {
		in, out := &in.AWSNodes, &out.AWSNodes
		*out = make([]AWSNode, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.LastASGModifiedTime != nil {
		in, out := &in.LastASGModifiedTime, &out.LastASGModifiedTime
		*out = (*in).DeepCopy()
	}
	if in.NextUpdateTime != nil {
		in, out := &in.NextUpdateTime, &out.NextUpdateTime
		*out = (*in).DeepCopy()
	}
	if in.UpdateStartTime != nil {
		in, out := &in.UpdateStartTime, &out.UpdateStartTime
		*out = (*in).DeepCopy()
	}
	if in.ReplaceTargetNode != nil {
		in, out := &in.ReplaceTargetNode, &out.ReplaceTargetNode
		*out = new(AWSNode)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNodeRefresherStatus.
func (in *AWSNodeRefresherStatus) DeepCopy() *AWSNodeRefresherStatus {
	if in == nil {
		return nil
	}
	out := new(AWSNodeRefresherStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNodeReplenisher) DeepCopyInto(out *AWSNodeReplenisher) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNodeReplenisher.
func (in *AWSNodeReplenisher) DeepCopy() *AWSNodeReplenisher {
	if in == nil {
		return nil
	}
	out := new(AWSNodeReplenisher)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSNodeReplenisher) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNodeReplenisherList) DeepCopyInto(out *AWSNodeReplenisherList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSNodeReplenisher, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNodeReplenisherList.
func (in *AWSNodeReplenisherList) DeepCopy() *AWSNodeReplenisherList {
	if in == nil {
		return nil
	}
	out := new(AWSNodeReplenisherList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSNodeReplenisherList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNodeReplenisherRef) DeepCopyInto(out *AWSNodeReplenisherRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNodeReplenisherRef.
func (in *AWSNodeReplenisherRef) DeepCopy() *AWSNodeReplenisherRef {
	if in == nil {
		return nil
	}
	out := new(AWSNodeReplenisherRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNodeReplenisherSpec) DeepCopyInto(out *AWSNodeReplenisherSpec) {
	*out = *in
	if in.AutoScalingGroups != nil {
		in, out := &in.AutoScalingGroups, &out.AutoScalingGroups
		*out = make([]AutoScalingGroup, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNodeReplenisherSpec.
func (in *AWSNodeReplenisherSpec) DeepCopy() *AWSNodeReplenisherSpec {
	if in == nil {
		return nil
	}
	out := new(AWSNodeReplenisherSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSNodeReplenisherStatus) DeepCopyInto(out *AWSNodeReplenisherStatus) {
	*out = *in
	if in.AWSNodes != nil {
		in, out := &in.AWSNodes, &out.AWSNodes
		*out = make([]AWSNode, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.NotJoinedAWSNodes != nil {
		in, out := &in.NotJoinedAWSNodes, &out.NotJoinedAWSNodes
		*out = make([]AWSNode, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.LastASGModifiedTime != nil {
		in, out := &in.LastASGModifiedTime, &out.LastASGModifiedTime
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSNodeReplenisherStatus.
func (in *AWSNodeReplenisherStatus) DeepCopy() *AWSNodeReplenisherStatus {
	if in == nil {
		return nil
	}
	out := new(AWSNodeReplenisherStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AutoScalingGroup) DeepCopyInto(out *AutoScalingGroup) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AutoScalingGroup.
func (in *AutoScalingGroup) DeepCopy() *AutoScalingGroup {
	if in == nil {
		return nil
	}
	out := new(AutoScalingGroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudAWS) DeepCopyInto(out *CloudAWS) {
	*out = *in
	if in.Masters != nil {
		in, out := &in.Masters, &out.Masters
		*out = new(Nodes)
		(*in).DeepCopyInto(*out)
	}
	if in.Workers != nil {
		in, out := &in.Workers, &out.Workers
		*out = new(Nodes)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudAWS.
func (in *CloudAWS) DeepCopy() *CloudAWS {
	if in == nil {
		return nil
	}
	out := new(CloudAWS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeManager) DeepCopyInto(out *NodeManager) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeManager.
func (in *NodeManager) DeepCopy() *NodeManager {
	if in == nil {
		return nil
	}
	out := new(NodeManager)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeManager) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeManagerList) DeepCopyInto(out *NodeManagerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NodeManager, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeManagerList.
func (in *NodeManagerList) DeepCopy() *NodeManagerList {
	if in == nil {
		return nil
	}
	out := new(NodeManagerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeManagerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeManagerSpec) DeepCopyInto(out *NodeManagerSpec) {
	*out = *in
	if in.Aws != nil {
		in, out := &in.Aws, &out.Aws
		*out = new(CloudAWS)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeManagerSpec.
func (in *NodeManagerSpec) DeepCopy() *NodeManagerSpec {
	if in == nil {
		return nil
	}
	out := new(NodeManagerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeManagerStatus) DeepCopyInto(out *NodeManagerStatus) {
	*out = *in
	if in.MasterAWSNodeManager != nil {
		in, out := &in.MasterAWSNodeManager, &out.MasterAWSNodeManager
		*out = new(AWSNodeManagerRef)
		**out = **in
	}
	if in.WorkerAWSNodeManager != nil {
		in, out := &in.WorkerAWSNodeManager, &out.WorkerAWSNodeManager
		*out = new(AWSNodeManagerRef)
		**out = **in
	}
	if in.MasterNodes != nil {
		in, out := &in.MasterNodes, &out.MasterNodes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.WorkerNodes != nil {
		in, out := &in.WorkerNodes, &out.WorkerNodes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeManagerStatus.
func (in *NodeManagerStatus) DeepCopy() *NodeManagerStatus {
	if in == nil {
		return nil
	}
	out := new(NodeManagerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Nodes) DeepCopyInto(out *Nodes) {
	*out = *in
	if in.AutoScalingGroups != nil {
		in, out := &in.AutoScalingGroups, &out.AutoScalingGroups
		*out = make([]AutoScalingGroup, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Nodes.
func (in *Nodes) DeepCopy() *Nodes {
	if in == nil {
		return nil
	}
	out := new(Nodes)
	in.DeepCopyInto(out)
	return out
}
