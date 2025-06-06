// Code generated by protoc-gen-go-helpers. DO NOT EDIT.
package persistence

import (
	"google.golang.org/protobuf/proto"
)

// Marshal an object of type BuildId to the protobuf v3 wire format
func (val *BuildId) Marshal() ([]byte, error) {
	return proto.Marshal(val)
}

// Unmarshal an object of type BuildId from the protobuf v3 wire format
func (val *BuildId) Unmarshal(buf []byte) error {
	return proto.Unmarshal(buf, val)
}

// Size returns the size of the object, in bytes, once serialized
func (val *BuildId) Size() int {
	return proto.Size(val)
}

// Equal returns whether two BuildId values are equivalent by recursively
// comparing the message's fields.
// For more information see the documentation for
// https://pkg.go.dev/google.golang.org/protobuf/proto#Equal
func (this *BuildId) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	var that1 *BuildId
	switch t := that.(type) {
	case *BuildId:
		that1 = t
	case BuildId:
		that1 = &t
	default:
		return false
	}

	return proto.Equal(this, that1)
}

// Marshal an object of type CompatibleVersionSet to the protobuf v3 wire format
func (val *CompatibleVersionSet) Marshal() ([]byte, error) {
	return proto.Marshal(val)
}

// Unmarshal an object of type CompatibleVersionSet from the protobuf v3 wire format
func (val *CompatibleVersionSet) Unmarshal(buf []byte) error {
	return proto.Unmarshal(buf, val)
}

// Size returns the size of the object, in bytes, once serialized
func (val *CompatibleVersionSet) Size() int {
	return proto.Size(val)
}

// Equal returns whether two CompatibleVersionSet values are equivalent by recursively
// comparing the message's fields.
// For more information see the documentation for
// https://pkg.go.dev/google.golang.org/protobuf/proto#Equal
func (this *CompatibleVersionSet) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	var that1 *CompatibleVersionSet
	switch t := that.(type) {
	case *CompatibleVersionSet:
		that1 = t
	case CompatibleVersionSet:
		that1 = &t
	default:
		return false
	}

	return proto.Equal(this, that1)
}

// Marshal an object of type AssignmentRule to the protobuf v3 wire format
func (val *AssignmentRule) Marshal() ([]byte, error) {
	return proto.Marshal(val)
}

// Unmarshal an object of type AssignmentRule from the protobuf v3 wire format
func (val *AssignmentRule) Unmarshal(buf []byte) error {
	return proto.Unmarshal(buf, val)
}

// Size returns the size of the object, in bytes, once serialized
func (val *AssignmentRule) Size() int {
	return proto.Size(val)
}

// Equal returns whether two AssignmentRule values are equivalent by recursively
// comparing the message's fields.
// For more information see the documentation for
// https://pkg.go.dev/google.golang.org/protobuf/proto#Equal
func (this *AssignmentRule) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	var that1 *AssignmentRule
	switch t := that.(type) {
	case *AssignmentRule:
		that1 = t
	case AssignmentRule:
		that1 = &t
	default:
		return false
	}

	return proto.Equal(this, that1)
}

// Marshal an object of type RedirectRule to the protobuf v3 wire format
func (val *RedirectRule) Marshal() ([]byte, error) {
	return proto.Marshal(val)
}

// Unmarshal an object of type RedirectRule from the protobuf v3 wire format
func (val *RedirectRule) Unmarshal(buf []byte) error {
	return proto.Unmarshal(buf, val)
}

// Size returns the size of the object, in bytes, once serialized
func (val *RedirectRule) Size() int {
	return proto.Size(val)
}

// Equal returns whether two RedirectRule values are equivalent by recursively
// comparing the message's fields.
// For more information see the documentation for
// https://pkg.go.dev/google.golang.org/protobuf/proto#Equal
func (this *RedirectRule) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	var that1 *RedirectRule
	switch t := that.(type) {
	case *RedirectRule:
		that1 = t
	case RedirectRule:
		that1 = &t
	default:
		return false
	}

	return proto.Equal(this, that1)
}

// Marshal an object of type VersioningData to the protobuf v3 wire format
func (val *VersioningData) Marshal() ([]byte, error) {
	return proto.Marshal(val)
}

// Unmarshal an object of type VersioningData from the protobuf v3 wire format
func (val *VersioningData) Unmarshal(buf []byte) error {
	return proto.Unmarshal(buf, val)
}

// Size returns the size of the object, in bytes, once serialized
func (val *VersioningData) Size() int {
	return proto.Size(val)
}

// Equal returns whether two VersioningData values are equivalent by recursively
// comparing the message's fields.
// For more information see the documentation for
// https://pkg.go.dev/google.golang.org/protobuf/proto#Equal
func (this *VersioningData) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	var that1 *VersioningData
	switch t := that.(type) {
	case *VersioningData:
		that1 = t
	case VersioningData:
		that1 = &t
	default:
		return false
	}

	return proto.Equal(this, that1)
}

// Marshal an object of type DeploymentData to the protobuf v3 wire format
func (val *DeploymentData) Marshal() ([]byte, error) {
	return proto.Marshal(val)
}

// Unmarshal an object of type DeploymentData from the protobuf v3 wire format
func (val *DeploymentData) Unmarshal(buf []byte) error {
	return proto.Unmarshal(buf, val)
}

// Size returns the size of the object, in bytes, once serialized
func (val *DeploymentData) Size() int {
	return proto.Size(val)
}

// Equal returns whether two DeploymentData values are equivalent by recursively
// comparing the message's fields.
// For more information see the documentation for
// https://pkg.go.dev/google.golang.org/protobuf/proto#Equal
func (this *DeploymentData) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	var that1 *DeploymentData
	switch t := that.(type) {
	case *DeploymentData:
		that1 = t
	case DeploymentData:
		that1 = &t
	default:
		return false
	}

	return proto.Equal(this, that1)
}

// Marshal an object of type TaskQueueTypeUserData to the protobuf v3 wire format
func (val *TaskQueueTypeUserData) Marshal() ([]byte, error) {
	return proto.Marshal(val)
}

// Unmarshal an object of type TaskQueueTypeUserData from the protobuf v3 wire format
func (val *TaskQueueTypeUserData) Unmarshal(buf []byte) error {
	return proto.Unmarshal(buf, val)
}

// Size returns the size of the object, in bytes, once serialized
func (val *TaskQueueTypeUserData) Size() int {
	return proto.Size(val)
}

// Equal returns whether two TaskQueueTypeUserData values are equivalent by recursively
// comparing the message's fields.
// For more information see the documentation for
// https://pkg.go.dev/google.golang.org/protobuf/proto#Equal
func (this *TaskQueueTypeUserData) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	var that1 *TaskQueueTypeUserData
	switch t := that.(type) {
	case *TaskQueueTypeUserData:
		that1 = t
	case TaskQueueTypeUserData:
		that1 = &t
	default:
		return false
	}

	return proto.Equal(this, that1)
}

// Marshal an object of type TaskQueueUserData to the protobuf v3 wire format
func (val *TaskQueueUserData) Marshal() ([]byte, error) {
	return proto.Marshal(val)
}

// Unmarshal an object of type TaskQueueUserData from the protobuf v3 wire format
func (val *TaskQueueUserData) Unmarshal(buf []byte) error {
	return proto.Unmarshal(buf, val)
}

// Size returns the size of the object, in bytes, once serialized
func (val *TaskQueueUserData) Size() int {
	return proto.Size(val)
}

// Equal returns whether two TaskQueueUserData values are equivalent by recursively
// comparing the message's fields.
// For more information see the documentation for
// https://pkg.go.dev/google.golang.org/protobuf/proto#Equal
func (this *TaskQueueUserData) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	var that1 *TaskQueueUserData
	switch t := that.(type) {
	case *TaskQueueUserData:
		that1 = t
	case TaskQueueUserData:
		that1 = &t
	default:
		return false
	}

	return proto.Equal(this, that1)
}

// Marshal an object of type VersionedTaskQueueUserData to the protobuf v3 wire format
func (val *VersionedTaskQueueUserData) Marshal() ([]byte, error) {
	return proto.Marshal(val)
}

// Unmarshal an object of type VersionedTaskQueueUserData from the protobuf v3 wire format
func (val *VersionedTaskQueueUserData) Unmarshal(buf []byte) error {
	return proto.Unmarshal(buf, val)
}

// Size returns the size of the object, in bytes, once serialized
func (val *VersionedTaskQueueUserData) Size() int {
	return proto.Size(val)
}

// Equal returns whether two VersionedTaskQueueUserData values are equivalent by recursively
// comparing the message's fields.
// For more information see the documentation for
// https://pkg.go.dev/google.golang.org/protobuf/proto#Equal
func (this *VersionedTaskQueueUserData) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	var that1 *VersionedTaskQueueUserData
	switch t := that.(type) {
	case *VersionedTaskQueueUserData:
		that1 = t
	case VersionedTaskQueueUserData:
		that1 = &t
	default:
		return false
	}

	return proto.Equal(this, that1)
}
