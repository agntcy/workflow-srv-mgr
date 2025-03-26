// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
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

// checks if the AgentMetadata type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &AgentMetadata{}

// AgentMetadata Basic information associated to the agent
type AgentMetadata struct {
	Ref AgentReference `json:"ref"`
	// Description of this agent, which should include what the intended use is, what tasks it accomplishes and how uses input and configs to produce the output and any other side effect
	Description string `json:"description"`
}

// NewAgentMetadata instantiates a new AgentMetadata object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewAgentMetadata(ref AgentReference, description string) *AgentMetadata {
	this := AgentMetadata{}
	this.Ref = ref
	this.Description = description
	return &this
}

// NewAgentMetadataWithDefaults instantiates a new AgentMetadata object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewAgentMetadataWithDefaults() *AgentMetadata {
	this := AgentMetadata{}
	return &this
}

// GetRef returns the Ref field value
func (o *AgentMetadata) GetRef() AgentReference {
	if o == nil {
		var ret AgentReference
		return ret
	}

	return o.Ref
}

// GetRefOk returns a tuple with the Ref field value
// and a boolean to check if the value has been set.
func (o *AgentMetadata) GetRefOk() (*AgentReference, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Ref, true
}

// SetRef sets field value
func (o *AgentMetadata) SetRef(v AgentReference) {
	o.Ref = v
}

// GetDescription returns the Description field value
func (o *AgentMetadata) GetDescription() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Description
}

// GetDescriptionOk returns a tuple with the Description field value
// and a boolean to check if the value has been set.
func (o *AgentMetadata) GetDescriptionOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Description, true
}

// SetDescription sets field value
func (o *AgentMetadata) SetDescription(v string) {
	o.Description = v
}

func (o AgentMetadata) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o AgentMetadata) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["ref"] = o.Ref
	toSerialize["description"] = o.Description
	return toSerialize, nil
}

type NullableAgentMetadata struct {
	value *AgentMetadata
	isSet bool
}

func (v NullableAgentMetadata) Get() *AgentMetadata {
	return v.value
}

func (v *NullableAgentMetadata) Set(val *AgentMetadata) {
	v.value = val
	v.isSet = true
}

func (v NullableAgentMetadata) IsSet() bool {
	return v.isSet
}

func (v *NullableAgentMetadata) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableAgentMetadata(val *AgentMetadata) *NullableAgentMetadata {
	return &NullableAgentMetadata{value: val, isSet: true}
}

func (v NullableAgentMetadata) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableAgentMetadata) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
