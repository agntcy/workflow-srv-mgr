// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
/*
Agent Manifest Definition

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: 0.2
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package manifests

import (
	"encoding/json"
)

// checks if the DockerDeployment type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &DockerDeployment{}

// DockerDeployment Describes the docker deployment for this agent
type DockerDeployment struct {
	Type string `json:"type"`
	// Name this deployment option is referred to within this agent. This is needed to indicate which one is preferred when this manifest is referred. Can be omitted, in such case selection is not possible.            -
	Name *string `json:"name,omitempty"`
	// Container image built for the agent containing the agent and Workflow Server.
	Image string `json:"image"`
}

// NewDockerDeployment instantiates a new DockerDeployment object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewDockerDeployment(type_ string, image string) *DockerDeployment {
	this := DockerDeployment{}
	this.Type = type_
	this.Image = image
	return &this
}

// NewDockerDeploymentWithDefaults instantiates a new DockerDeployment object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewDockerDeploymentWithDefaults() *DockerDeployment {
	this := DockerDeployment{}
	return &this
}

// GetType returns the Type field value
func (o *DockerDeployment) GetType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Type
}

// GetTypeOk returns a tuple with the Type field value
// and a boolean to check if the value has been set.
func (o *DockerDeployment) GetTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Type, true
}

// SetType sets field value
func (o *DockerDeployment) SetType(v string) {
	o.Type = v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *DockerDeployment) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *DockerDeployment) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *DockerDeployment) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *DockerDeployment) SetName(v string) {
	o.Name = &v
}

// GetImage returns the Image field value
func (o *DockerDeployment) GetImage() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Image
}

// GetImageOk returns a tuple with the Image field value
// and a boolean to check if the value has been set.
func (o *DockerDeployment) GetImageOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Image, true
}

// SetImage sets field value
func (o *DockerDeployment) SetImage(v string) {
	o.Image = v
}

func (o DockerDeployment) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o DockerDeployment) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["type"] = o.Type
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	toSerialize["image"] = o.Image
	return toSerialize, nil
}

type NullableDockerDeployment struct {
	value *DockerDeployment
	isSet bool
}

func (v NullableDockerDeployment) Get() *DockerDeployment {
	return v.value
}

func (v *NullableDockerDeployment) Set(val *DockerDeployment) {
	v.value = val
	v.isSet = true
}

func (v NullableDockerDeployment) IsSet() bool {
	return v.isSet
}

func (v *NullableDockerDeployment) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableDockerDeployment(val *DockerDeployment) *NullableDockerDeployment {
	return &NullableDockerDeployment{value: val, isSet: true}
}

func (v NullableDockerDeployment) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableDockerDeployment) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
