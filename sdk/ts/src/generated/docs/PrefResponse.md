# PrefResponse

Preference response payload 

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**user_id** | **string** | Unique identifier of the user  | [optional] [default to undefined]
**app_id** | **string** | Unique identifier of the application  | [optional] [default to undefined]
**updated_at** | **string** | Timestamp of last update  | [optional] [default to undefined]
**data** | **any** | Arbitrary JSON preference payload.  | [optional] [default to undefined]

## Example

```typescript
import { PrefResponse } from './api';

const instance: PrefResponse = {
    user_id,
    app_id,
    updated_at,
    data,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
