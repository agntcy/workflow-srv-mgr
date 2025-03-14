/*
Agent Manifest Definition

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: 0.1
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
	// Container image for the agent
	Image    string               `json:"image"`
	Protocol AgentConnectProtocol `json:"protocol"`
}

// NewDockerDeployment instantiates a new DockerDeployment object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewDockerDeployment(type_ string, image string, protocol AgentConnectProtocol) *DockerDeployment {
	this := DockerDeployment{}
	this.Type = type_
	this.Image = image
	this.Protocol = protocol
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

// GetProtocol returns the Protocol field value
func (o *DockerDeployment) GetProtocol() AgentConnectProtocol {
	if o == nil {
		var ret AgentConnectProtocol
		return ret
	}

	return o.Protocol
}

// GetProtocolOk returns a tuple with the Protocol field value
// and a boolean to check if the value has been set.
func (o *DockerDeployment) GetProtocolOk() (*AgentConnectProtocol, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Protocol, true
}

// SetProtocol sets field value
func (o *DockerDeployment) SetProtocol(v AgentConnectProtocol) {
	o.Protocol = v
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
	toSerialize["image"] = o.Image
	toSerialize["protocol"] = o.Protocol
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
