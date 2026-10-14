# AppResponse

Application response payload 

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | Unique identifier of the application  | [optional] [default to undefined]
**name** | **string** | Name of the application  | [optional] [default to undefined]
**desc** | **string** | Description of the application  | [optional] [default to undefined]
**max_payload** | **number** | Maximum payload size in bytes  | [optional] [default to undefined]
**max_user** | **number** | Maximum number of users  | [optional] [default to undefined]
**created_by** | **string** | Unique identifier of the user who created the application  | [optional] [default to undefined]

## Example

```typescript
import { AppResponse } from './api';

const instance: AppResponse = {
    id,
    name,
    desc,
    max_payload,
    max_user,
    created_by,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
