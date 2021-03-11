/*
 * HighLoad Cup 2021
 *
 * ## Usage ## List of all custom errors First number is HTTP Status code, second is value of \"code\" field in returned JSON object, text description may or may not match \"message\" field in returned JSON object. - 422.1000: wrong coordinates - 422.1001: wrong depth - 409.1002: no more active licenses allowed - 409.1003: treasure is not digged 
 *
 * API version: 1.0.0
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
)

// License License for digging.
type License struct {
	Id int32 `json:"id"`
	// Non-negative amount of treasures/etc.
	DigAllowed int32 `json:"digAllowed"`
	// Non-negative amount of treasures/etc.
	DigUsed int32 `json:"digUsed"`
}

// NewLicense instantiates a new License object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewLicense(id int32, digAllowed int32, digUsed int32) *License {
	this := License{}
	this.Id = id
	this.DigAllowed = digAllowed
	this.DigUsed = digUsed
	return &this
}

// NewLicenseWithDefaults instantiates a new License object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewLicenseWithDefaults() *License {
	this := License{}
	return &this
}

// GetId returns the Id field value
func (o *License) GetId() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *License) GetIdOk() (*int32, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *License) SetId(v int32) {
	o.Id = v
}

// GetDigAllowed returns the DigAllowed field value
func (o *License) GetDigAllowed() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.DigAllowed
}

// GetDigAllowedOk returns a tuple with the DigAllowed field value
// and a boolean to check if the value has been set.
func (o *License) GetDigAllowedOk() (*int32, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.DigAllowed, true
}

// SetDigAllowed sets field value
func (o *License) SetDigAllowed(v int32) {
	o.DigAllowed = v
}

// GetDigUsed returns the DigUsed field value
func (o *License) GetDigUsed() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.DigUsed
}

// GetDigUsedOk returns a tuple with the DigUsed field value
// and a boolean to check if the value has been set.
func (o *License) GetDigUsedOk() (*int32, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.DigUsed, true
}

// SetDigUsed sets field value
func (o *License) SetDigUsed(v int32) {
	o.DigUsed = v
}

func (o License) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["id"] = o.Id
	}
	if true {
		toSerialize["digAllowed"] = o.DigAllowed
	}
	if true {
		toSerialize["digUsed"] = o.DigUsed
	}
	return json.Marshal(toSerialize)
}

type NullableLicense struct {
	value *License
	isSet bool
}

func (v NullableLicense) Get() *License {
	return v.value
}

func (v *NullableLicense) Set(val *License) {
	v.value = val
	v.isSet = true
}

func (v NullableLicense) IsSet() bool {
	return v.isSet
}

func (v *NullableLicense) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableLicense(val *License) *NullableLicense {
	return &NullableLicense{value: val, isSet: true}
}

func (v NullableLicense) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableLicense) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


