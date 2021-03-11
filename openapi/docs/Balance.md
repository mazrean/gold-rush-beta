# Balance

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Balance** | **int32** |  | 
**Wallet** | **[]int32** | Wallet with some coins. | 

## Methods

### NewBalance

`func NewBalance(balance int32, wallet []int32, ) *Balance`

NewBalance instantiates a new Balance object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBalanceWithDefaults

`func NewBalanceWithDefaults() *Balance`

NewBalanceWithDefaults instantiates a new Balance object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBalance

`func (o *Balance) GetBalance() int32`

GetBalance returns the Balance field if non-nil, zero value otherwise.

### GetBalanceOk

`func (o *Balance) GetBalanceOk() (*int32, bool)`

GetBalanceOk returns a tuple with the Balance field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBalance

`func (o *Balance) SetBalance(v int32)`

SetBalance sets Balance field to given value.


### GetWallet

`func (o *Balance) GetWallet() []int32`

GetWallet returns the Wallet field if non-nil, zero value otherwise.

### GetWalletOk

`func (o *Balance) GetWalletOk() (*[]int32, bool)`

GetWalletOk returns a tuple with the Wallet field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWallet

`func (o *Balance) SetWallet(v []int32)`

SetWallet sets Wallet field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


