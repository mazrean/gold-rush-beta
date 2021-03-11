# Report

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Area** | [**Area**](area.md) |  | 
**Amount** | **int32** | Non-negative amount of treasures/etc. | 

## Methods

### NewReport

`func NewReport(area Area, amount int32, ) *Report`

NewReport instantiates a new Report object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReportWithDefaults

`func NewReportWithDefaults() *Report`

NewReportWithDefaults instantiates a new Report object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetArea

`func (o *Report) GetArea() Area`

GetArea returns the Area field if non-nil, zero value otherwise.

### GetAreaOk

`func (o *Report) GetAreaOk() (*Area, bool)`

GetAreaOk returns a tuple with the Area field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArea

`func (o *Report) SetArea(v Area)`

SetArea sets Area field to given value.


### GetAmount

`func (o *Report) GetAmount() int32`

GetAmount returns the Amount field if non-nil, zero value otherwise.

### GetAmountOk

`func (o *Report) GetAmountOk() (*int32, bool)`

GetAmountOk returns a tuple with the Amount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAmount

`func (o *Report) SetAmount(v int32)`

SetAmount sets Amount field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


