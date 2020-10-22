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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Actor) DeepCopyInto(out *Actor) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Actor.
func (in *Actor) DeepCopy() *Actor {
	if in == nil {
		return nil
	}
	out := new(Actor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Director) DeepCopyInto(out *Director) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Director.
func (in *Director) DeepCopy() *Director {
	if in == nil {
		return nil
	}
	out := new(Director)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProblemSettings) DeepCopyInto(out *ProblemSettings) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProblemSettings.
func (in *ProblemSettings) DeepCopy() *ProblemSettings {
	if in == nil {
		return nil
	}
	out := new(ProblemSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Room) DeepCopyInto(out *Room) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Room.
func (in *Room) DeepCopy() *Room {
	if in == nil {
		return nil
	}
	out := new(Room)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Room) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoomList) DeepCopyInto(out *RoomList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Room, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoomList.
func (in *RoomList) DeepCopy() *RoomList {
	if in == nil {
		return nil
	}
	out := new(RoomList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RoomList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoomSpec) DeepCopyInto(out *RoomSpec) {
	*out = *in
	if in.ProblemSettings != nil {
		in, out := &in.ProblemSettings, &out.ProblemSettings
		*out = new(ProblemSettings)
		**out = **in
	}
	if in.Actors != nil {
		in, out := &in.Actors, &out.Actors
		*out = make([]Actor, len(*in))
		copy(*out, *in)
	}
	out.Director = in.Director
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoomSpec.
func (in *RoomSpec) DeepCopy() *RoomSpec {
	if in == nil {
		return nil
	}
	out := new(RoomSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoomStatus) DeepCopyInto(out *RoomStatus) {
	*out = *in
	if in.GimulatorStatus != nil {
		in, out := &in.GimulatorStatus, &out.GimulatorStatus
		*out = new(corev1.PodStatus)
		(*in).DeepCopyInto(*out)
	}
	if in.DirectorStatus != nil {
		in, out := &in.DirectorStatus, &out.DirectorStatus
		*out = new(corev1.PodStatus)
		(*in).DeepCopyInto(*out)
	}
	if in.ActorStatuses != nil {
		in, out := &in.ActorStatuses, &out.ActorStatuses
		*out = make(map[string]*corev1.PodStatus, len(*in))
		for key, val := range *in {
			var outVal *corev1.PodStatus
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(corev1.PodStatus)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoomStatus.
func (in *RoomStatus) DeepCopy() *RoomStatus {
	if in == nil {
		return nil
	}
	out := new(RoomStatus)
	in.DeepCopyInto(out)
	return out
}
