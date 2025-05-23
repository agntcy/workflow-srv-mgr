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

// checks if the Locator type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &Locator{}

// Locator struct for Locator
type Locator struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Digest      *string           `json:"digest,omitempty"`
	Size        *int32            `json:"size,omitempty"`
	Type        string            `json:"type"`
	Url         string            `json:"url"`
}

// NewLocator instantiates a new Locator object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewLocator(type_ string, url string) *Locator {
	this := Locator{}
	this.Type = type_
	this.Url = url
	return &this
}

// NewLocatorWithDefaults instantiates a new Locator object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewLocatorWithDefaults() *Locator {
	this := Locator{}
	return &this
}

// GetAnnotations returns the Annotations field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *Locator) GetAnnotations() map[string]string {
	if o == nil {
		var ret map[string]string
		return ret
	}
	return o.Annotations
}

// GetAnnotationsOk returns a tuple with the Annotations field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *Locator) GetAnnotationsOk() (*map[string]string, bool) {
	if o == nil || IsNil(o.Annotations) {
		return nil, false
	}
	return &o.Annotations, true
}

// HasAnnotations returns a boolean if a field has been set.
func (o *Locator) HasAnnotations() bool {
	if o != nil && IsNil(o.Annotations) {
		return true
	}

	return false
}

// SetAnnotations gets a reference to the given map[string]string and assigns it to the Annotations field.
func (o *Locator) SetAnnotations(v map[string]string) {
	o.Annotations = v
}

// GetDigest returns the Digest field value if set, zero value otherwise.
func (o *Locator) GetDigest() string {
	if o == nil || IsNil(o.Digest) {
		var ret string
		return ret
	}
	return *o.Digest
}

// GetDigestOk returns a tuple with the Digest field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Locator) GetDigestOk() (*string, bool) {
	if o == nil || IsNil(o.Digest) {
		return nil, false
	}
	return o.Digest, true
}

// HasDigest returns a boolean if a field has been set.
func (o *Locator) HasDigest() bool {
	if o != nil && !IsNil(o.Digest) {
		return true
	}

	return false
}

// SetDigest gets a reference to the given string and assigns it to the Digest field.
func (o *Locator) SetDigest(v string) {
	o.Digest = &v
}

// GetSize returns the Size field value if set, zero value otherwise.
func (o *Locator) GetSize() int32 {
	if o == nil || IsNil(o.Size) {
		var ret int32
		return ret
	}
	return *o.Size
}

// GetSizeOk returns a tuple with the Size field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Locator) GetSizeOk() (*int32, bool) {
	if o == nil || IsNil(o.Size) {
		return nil, false
	}
	return o.Size, true
}

// HasSize returns a boolean if a field has been set.
func (o *Locator) HasSize() bool {
	if o != nil && !IsNil(o.Size) {
		return true
	}

	return false
}

// SetSize gets a reference to the given int32 and assigns it to the Size field.
func (o *Locator) SetSize(v int32) {
	o.Size = &v
}

// GetType returns the Type field value
func (o *Locator) GetType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Type
}

// GetTypeOk returns a tuple with the Type field value
// and a boolean to check if the value has been set.
func (o *Locator) GetTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Type, true
}

// SetType sets field value
func (o *Locator) SetType(v string) {
	o.Type = v
}

// GetUrl returns the Url field value
func (o *Locator) GetUrl() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Url
}

// GetUrlOk returns a tuple with the Url field value
// and a boolean to check if the value has been set.
func (o *Locator) GetUrlOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Url, true
}

// SetUrl sets field value
func (o *Locator) SetUrl(v string) {
	o.Url = v
}

func (o Locator) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o Locator) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if o.Annotations != nil {
		toSerialize["annotations"] = o.Annotations
	}
	if !IsNil(o.Digest) {
		toSerialize["digest"] = o.Digest
	}
	if !IsNil(o.Size) {
		toSerialize["size"] = o.Size
	}
	toSerialize["type"] = o.Type
	toSerialize["url"] = o.Url
	return toSerialize, nil
}

type NullableLocator struct {
	value *Locator
	isSet bool
}

func (v NullableLocator) Get() *Locator {
	return v.value
}

func (v *NullableLocator) Set(val *Locator) {
	v.value = val
	v.isSet = true
}

func (v NullableLocator) IsSet() bool {
	return v.isSet
}

func (v *NullableLocator) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableLocator(val *Locator) *NullableLocator {
	return &NullableLocator{value: val, isSet: true}
}

func (v NullableLocator) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableLocator) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
