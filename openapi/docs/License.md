# License

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **int32** |  | 
**DigAllowed** | **int32** | Non-negative amount of treasures/etc. | 
**DigUsed** | **int32** | Non-negative amount of treasures/etc. | 

## Methods

### NewLicense

`func NewLicense(id int32, digAllowed int32, digUsed int32, ) *License`

NewLicense instantiates a new License object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLicenseWithDefaults

`func NewLicenseWithDefaults() *License`

NewLicenseWithDefaults instantiates a new License object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *License) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *License) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *License) SetId(v int32)`

SetId sets Id field to given value.


### GetDigAllowed

`func (o *License) GetDigAllowed() int32`

GetDigAllowed returns the DigAllowed field if non-nil, zero value otherwise.

### GetDigAllowedOk

`func (o *License) GetDigAllowedOk() (*int32, bool)`

GetDigAllowedOk returns a tuple with the DigAllowed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDigAllowed

`func (o *License) SetDigAllowed(v int32)`

SetDigAllowed sets DigAllowed field to given value.


### GetDigUsed

`func (o *License) GetDigUsed() int32`

GetDigUsed returns the DigUsed field if non-nil, zero value otherwise.

### GetDigUsedOk

`func (o *License) GetDigUsedOk() (*int32, bool)`

GetDigUsedOk returns a tuple with the DigUsed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDigUsed

`func (o *License) SetDigUsed(v int32)`

SetDigUsed sets DigUsed field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


