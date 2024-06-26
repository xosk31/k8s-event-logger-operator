//go:build !ignore_autogenerated

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
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventLogger) DeepCopyInto(out *EventLogger) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventLogger.
func (in *EventLogger) DeepCopy() *EventLogger {
	if in == nil {
		return nil
	}
	out := new(EventLogger)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EventLogger) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventLoggerList) DeepCopyInto(out *EventLoggerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]EventLogger, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventLoggerList.
func (in *EventLoggerList) DeepCopy() *EventLoggerList {
	if in == nil {
		return nil
	}
	out := new(EventLoggerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EventLoggerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventLoggerSpec) DeepCopyInto(out *EventLoggerSpec) {
	*out = *in
	if in.Kinds != nil {
		in, out := &in.Kinds, &out.Kinds
		*out = make([]Kind, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.EventTypes != nil {
		in, out := &in.EventTypes, &out.EventTypes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.ScrapeMetrics != nil {
		in, out := &in.ScrapeMetrics, &out.ScrapeMetrics
		*out = new(bool)
		**out = **in
	}
	if in.Namespace != nil {
		in, out := &in.Namespace, &out.Namespace
		*out = new(string)
		**out = **in
	}
	if in.ImagePullSecrets != nil {
		in, out := &in.ImagePullSecrets, &out.ImagePullSecrets
		*out = make([]corev1.LocalObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.LogFields != nil {
		in, out := &in.LogFields, &out.LogFields
		*out = make([]LogField, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventLoggerSpec.
func (in *EventLoggerSpec) DeepCopy() *EventLoggerSpec {
	if in == nil {
		return nil
	}
	out := new(EventLoggerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventLoggerStatus) DeepCopyInto(out *EventLoggerStatus) {
	*out = *in
	in.LastProcessed.DeepCopyInto(&out.LastProcessed)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventLoggerStatus.
func (in *EventLoggerStatus) DeepCopy() *EventLoggerStatus {
	if in == nil {
		return nil
	}
	out := new(EventLoggerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Kind) DeepCopyInto(out *Kind) {
	*out = *in
	if in.APIGroup != nil {
		in, out := &in.APIGroup, &out.APIGroup
		*out = new(string)
		**out = **in
	}
	if in.EventTypes != nil {
		in, out := &in.EventTypes, &out.EventTypes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Reasons != nil {
		in, out := &in.Reasons, &out.Reasons
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SkipReasons != nil {
		in, out := &in.SkipReasons, &out.SkipReasons
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.MatchingPatterns != nil {
		in, out := &in.MatchingPatterns, &out.MatchingPatterns
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SkipOnMatch != nil {
		in, out := &in.SkipOnMatch, &out.SkipOnMatch
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Kind.
func (in *Kind) DeepCopy() *Kind {
	if in == nil {
		return nil
	}
	out := new(Kind)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogField) DeepCopyInto(out *LogField) {
	*out = *in
	if in.Path != nil {
		in, out := &in.Path, &out.Path
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogField.
func (in *LogField) DeepCopy() *LogField {
	if in == nil {
		return nil
	}
	out := new(LogField)
	in.DeepCopyInto(out)
	return out
}
