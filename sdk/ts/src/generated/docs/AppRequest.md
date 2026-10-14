# AppRequest

Payload for creating or updating an application 

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** | Name of the application  | [optional] [default to undefined]
**desc** | **string** | Description of the application  | [optional] [default to undefined]
**max_payload** | **number** | Maximum payload size in bytes  | [optional] [default to undefined]
**max_user** | **number** | Maximum number of users  | [optional] [default to undefined]

## Example

```typescript
import { AppRequest } from './api';

const instance: AppRequest = {
    name,
    desc,
    max_payload,
    max_user,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
