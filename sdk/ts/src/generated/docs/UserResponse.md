# UserResponse

User response payload 

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | Unique identifier of the user  | [optional] [default to undefined]
**auth_provider** | **string** | Authentication provider name  | [optional] [default to undefined]
**auth_sub** | **string** | Authentication subject identifier  | [optional] [default to undefined]
**apps** | [**Array&lt;AppResponse&gt;**](AppResponse.md) | List of applications the user is authorised for  | [optional] [default to undefined]

## Example

```typescript
import { UserResponse } from './api';

const instance: UserResponse = {
    id,
    auth_provider,
    auth_sub,
    apps,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
