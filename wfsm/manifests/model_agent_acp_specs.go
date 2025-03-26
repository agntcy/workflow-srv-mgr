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

// checks if the AgentACPSpecs type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &AgentACPSpecs{}

// AgentACPSpecs Specification of agent capabilities, config, input, output, and interrupts
type AgentACPSpecs struct {
	Capabilities AgentCapabilities `json:"capabilities"`
	// This object contains an instance of an OpenAPI schema object, formatted as per the OpenAPI specs: https://spec.openapis.org/oas/v3.1.1.html#schema-object
	Input map[string]interface{} `json:"input"`
	// This object contains an instance of an OpenAPI schema object, formatted as per the OpenAPI specs: https://spec.openapis.org/oas/v3.1.1.html#schema-object
	Output map[string]interface{} `json:"output"`
	// This describes the format of an Update in the streaming.  Must be specified if `streaming.custom` capability is true and cannot be specified otherwise. Format follows: https://spec.openapis.org/oas/v3.1.1.html#schema-object
	CustomStreamingUpdate map[string]interface{} `json:"custom_streaming_update,omitempty"`
	// This describes the format of ThreadState.  Cannot be specified if `threads` capability is false. If not specified, when `threads` capability is true, then the API to retrieve ThreadState from a Thread or a Run is not available. This object contains an instance of an OpenAPI schema object, formatted as per the OpenAPI specs: https://spec.openapis.org/oas/v3.1.1.html#schema-object
	ThreadState map[string]interface{} `json:"thread_state,omitempty"`
	// This object contains an instance of an OpenAPI schema object, formatted as per the OpenAPI specs: https://spec.openapis.org/oas/v3.1.1.html#schema-object
	Config     map[string]interface{}         `json:"config"`
	Interrupts []AgentACPSpecsInterruptsInner `json:"interrupts,omitempty"`
}

// NewAgentACPSpecs instantiates a new AgentACPSpecs object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewAgentACPSpecs(capabilities AgentCapabilities, input map[string]interface{}, output map[string]interface{}, config map[string]interface{}) *AgentACPSpecs {
	this := AgentACPSpecs{}
	this.Capabilities = capabilities
	this.Input = input
	this.Output = output
	this.Config = config
	return &this
}

// NewAgentACPSpecsWithDefaults instantiates a new AgentACPSpecs object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewAgentACPSpecsWithDefaults() *AgentACPSpecs {
	this := AgentACPSpecs{}
	return &this
}

// GetCapabilities returns the Capabilities field value
func (o *AgentACPSpecs) GetCapabilities() AgentCapabilities {
	if o == nil {
		var ret AgentCapabilities
		return ret
	}

	return o.Capabilities
}

// GetCapabilitiesOk returns a tuple with the Capabilities field value
// and a boolean to check if the value has been set.
func (o *AgentACPSpecs) GetCapabilitiesOk() (*AgentCapabilities, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Capabilities, true
}

// SetCapabilities sets field value
func (o *AgentACPSpecs) SetCapabilities(v AgentCapabilities) {
	o.Capabilities = v
}

// GetInput returns the Input field value
func (o *AgentACPSpecs) GetInput() map[string]interface{} {
	if o == nil {
		var ret map[string]interface{}
		return ret
	}

	return o.Input
}

// GetInputOk returns a tuple with the Input field value
// and a boolean to check if the value has been set.
func (o *AgentACPSpecs) GetInputOk() (map[string]interface{}, bool) {
	if o == nil {
		return map[string]interface{}{}, false
	}
	return o.Input, true
}

// SetInput sets field value
func (o *AgentACPSpecs) SetInput(v map[string]interface{}) {
	o.Input = v
}

// GetOutput returns the Output field value
func (o *AgentACPSpecs) GetOutput() map[string]interface{} {
	if o == nil {
		var ret map[string]interface{}
		return ret
	}

	return o.Output
}

// GetOutputOk returns a tuple with the Output field value
// and a boolean to check if the value has been set.
func (o *AgentACPSpecs) GetOutputOk() (map[string]interface{}, bool) {
	if o == nil {
		return map[string]interface{}{}, false
	}
	return o.Output, true
}

// SetOutput sets field value
func (o *AgentACPSpecs) SetOutput(v map[string]interface{}) {
	o.Output = v
}

// GetCustomStreamingUpdate returns the CustomStreamingUpdate field value if set, zero value otherwise.
func (o *AgentACPSpecs) GetCustomStreamingUpdate() map[string]interface{} {
	if o == nil || IsNil(o.CustomStreamingUpdate) {
		var ret map[string]interface{}
		return ret
	}
	return o.CustomStreamingUpdate
}

// GetCustomStreamingUpdateOk returns a tuple with the CustomStreamingUpdate field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AgentACPSpecs) GetCustomStreamingUpdateOk() (map[string]interface{}, bool) {
	if o == nil || IsNil(o.CustomStreamingUpdate) {
		return map[string]interface{}{}, false
	}
	return o.CustomStreamingUpdate, true
}

// HasCustomStreamingUpdate returns a boolean if a field has been set.
func (o *AgentACPSpecs) HasCustomStreamingUpdate() bool {
	if o != nil && !IsNil(o.CustomStreamingUpdate) {
		return true
	}

	return false
}

// SetCustomStreamingUpdate gets a reference to the given map[string]interface{} and assigns it to the CustomStreamingUpdate field.
func (o *AgentACPSpecs) SetCustomStreamingUpdate(v map[string]interface{}) {
	o.CustomStreamingUpdate = v
}

// GetThreadState returns the ThreadState field value if set, zero value otherwise.
func (o *AgentACPSpecs) GetThreadState() map[string]interface{} {
	if o == nil || IsNil(o.ThreadState) {
		var ret map[string]interface{}
		return ret
	}
	return o.ThreadState
}

// GetThreadStateOk returns a tuple with the ThreadState field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AgentACPSpecs) GetThreadStateOk() (map[string]interface{}, bool) {
	if o == nil || IsNil(o.ThreadState) {
		return map[string]interface{}{}, false
	}
	return o.ThreadState, true
}

// HasThreadState returns a boolean if a field has been set.
func (o *AgentACPSpecs) HasThreadState() bool {
	if o != nil && !IsNil(o.ThreadState) {
		return true
	}

	return false
}

// SetThreadState gets a reference to the given map[string]interface{} and assigns it to the ThreadState field.
func (o *AgentACPSpecs) SetThreadState(v map[string]interface{}) {
	o.ThreadState = v
}

// GetConfig returns the Config field value
func (o *AgentACPSpecs) GetConfig() map[string]interface{} {
	if o == nil {
		var ret map[string]interface{}
		return ret
	}

	return o.Config
}

// GetConfigOk returns a tuple with the Config field value
// and a boolean to check if the value has been set.
func (o *AgentACPSpecs) GetConfigOk() (map[string]interface{}, bool) {
	if o == nil {
		return map[string]interface{}{}, false
	}
	return o.Config, true
}

// SetConfig sets field value
func (o *AgentACPSpecs) SetConfig(v map[string]interface{}) {
	o.Config = v
}

// GetInterrupts returns the Interrupts field value if set, zero value otherwise.
func (o *AgentACPSpecs) GetInterrupts() []AgentACPSpecsInterruptsInner {
	if o == nil || IsNil(o.Interrupts) {
		var ret []AgentACPSpecsInterruptsInner
		return ret
	}
	return o.Interrupts
}

// GetInterruptsOk returns a tuple with the Interrupts field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AgentACPSpecs) GetInterruptsOk() ([]AgentACPSpecsInterruptsInner, bool) {
	if o == nil || IsNil(o.Interrupts) {
		return nil, false
	}
	return o.Interrupts, true
}

// HasInterrupts returns a boolean if a field has been set.
func (o *AgentACPSpecs) HasInterrupts() bool {
	if o != nil && !IsNil(o.Interrupts) {
		return true
	}

	return false
}

// SetInterrupts gets a reference to the given []AgentACPSpecsInterruptsInner and assigns it to the Interrupts field.
func (o *AgentACPSpecs) SetInterrupts(v []AgentACPSpecsInterruptsInner) {
	o.Interrupts = v
}

func (o AgentACPSpecs) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o AgentACPSpecs) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["capabilities"] = o.Capabilities
	toSerialize["input"] = o.Input
	toSerialize["output"] = o.Output
	if !IsNil(o.CustomStreamingUpdate) {
		toSerialize["custom_streaming_update"] = o.CustomStreamingUpdate
	}
	if !IsNil(o.ThreadState) {
		toSerialize["thread_state"] = o.ThreadState
	}
	toSerialize["config"] = o.Config
	if !IsNil(o.Interrupts) {
		toSerialize["interrupts"] = o.Interrupts
	}
	return toSerialize, nil
}

type NullableAgentACPSpecs struct {
	value *AgentACPSpecs
	isSet bool
}

func (v NullableAgentACPSpecs) Get() *AgentACPSpecs {
	return v.value
}

func (v *NullableAgentACPSpecs) Set(val *AgentACPSpecs) {
	v.value = val
	v.isSet = true
}

func (v NullableAgentACPSpecs) IsSet() bool {
	return v.isSet
}

func (v *NullableAgentACPSpecs) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableAgentACPSpecs(val *AgentACPSpecs) *NullableAgentACPSpecs {
	return &NullableAgentACPSpecs{value: val, isSet: true}
}

func (v NullableAgentACPSpecs) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableAgentACPSpecs) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
