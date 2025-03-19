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

// checks if the AgentReference type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &AgentReference{}

// AgentReference Reference to the agent in the agent directory. It includes the version and the locator.
type AgentReference struct {
	// Name of the agent that identifies the agent in its record
	Name string `json:"name"`
	// Version of the agent in its record. Should be formatted according to semantic versioning (https://semver.org)
	Version string `json:"version"`
	// URL of the record. Can be a network location or a file.
	Url *string `json:"url,omitempty"`
}

// NewAgentReference instantiates a new AgentReference object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewAgentReference(name string, version string) *AgentReference {
	this := AgentReference{}
	this.Name = name
	this.Version = version
	return &this
}

// NewAgentReferenceWithDefaults instantiates a new AgentReference object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewAgentReferenceWithDefaults() *AgentReference {
	this := AgentReference{}
	return &this
}

// GetName returns the Name field value
func (o *AgentReference) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *AgentReference) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *AgentReference) SetName(v string) {
	o.Name = v
}

// GetVersion returns the Version field value
func (o *AgentReference) GetVersion() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Version
}

// GetVersionOk returns a tuple with the Version field value
// and a boolean to check if the value has been set.
func (o *AgentReference) GetVersionOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Version, true
}

// SetVersion sets field value
func (o *AgentReference) SetVersion(v string) {
	o.Version = v
}

// GetUrl returns the Url field value if set, zero value otherwise.
func (o *AgentReference) GetUrl() string {
	if o == nil || IsNil(o.Url) {
		var ret string
		return ret
	}
	return *o.Url
}

// GetUrlOk returns a tuple with the Url field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AgentReference) GetUrlOk() (*string, bool) {
	if o == nil || IsNil(o.Url) {
		return nil, false
	}
	return o.Url, true
}

// HasUrl returns a boolean if a field has been set.
func (o *AgentReference) HasUrl() bool {
	if o != nil && !IsNil(o.Url) {
		return true
	}

	return false
}

// SetUrl gets a reference to the given string and assigns it to the Url field.
func (o *AgentReference) SetUrl(v string) {
	o.Url = &v
}

func (o AgentReference) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o AgentReference) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["name"] = o.Name
	toSerialize["version"] = o.Version
	if !IsNil(o.Url) {
		toSerialize["url"] = o.Url
	}
	return toSerialize, nil
}

type NullableAgentReference struct {
	value *AgentReference
	isSet bool
}

func (v NullableAgentReference) Get() *AgentReference {
	return v.value
}

func (v *NullableAgentReference) Set(val *AgentReference) {
	v.value = val
	v.isSet = true
}

func (v NullableAgentReference) IsSet() bool {
	return v.isSet
}

func (v *NullableAgentReference) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableAgentReference(val *AgentReference) *NullableAgentReference {
	return &NullableAgentReference{value: val, isSet: true}
}

func (v NullableAgentReference) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableAgentReference) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
