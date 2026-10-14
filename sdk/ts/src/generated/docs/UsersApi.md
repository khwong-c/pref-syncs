# UsersApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**authoriseUser**](#authoriseuser) | **PUT** /user/to/{app} | Authorise user to an app |
|[**deauthoriseUser**](#deauthoriseuser) | **DELETE** /user/from/{app} | Deauthorise user from an app |
|[**deleteUser**](#deleteuser) | **DELETE** /user | Delete current user |
|[**getUser**](#getuser) | **GET** /user | Get current user |

# **authoriseUser**
> UserResponse authoriseUser()

Grants the authenticated user access to the specified app. Requires an authenticated user. 

### Example

```typescript
import {
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let app: string; //The unique identifier of the app  (default to undefined)

const { status, data } = await apiInstance.authoriseUser(
    app
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **app** | [**string**] | The unique identifier of the app  | defaults to undefined|


### Return type

**UserResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | User authorised to application  |  -  |
|**0** | Error response  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deauthoriseUser**
> UserResponse deauthoriseUser()

Revokes the authenticated user\'s access to the specified app. Requires an authenticated user. 

### Example

```typescript
import {
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let app: string; //The unique identifier of the app  (default to undefined)

const { status, data } = await apiInstance.deauthoriseUser(
    app
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **app** | [**string**] | The unique identifier of the app  | defaults to undefined|


### Return type

**UserResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | User deauthorised from application  |  -  |
|**0** | Error response  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteUser**
> SuccessResponse deleteUser()

Permanently deletes the authenticated user\'s account. Requires an authenticated user. 

### Example

```typescript
import {
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

const { status, data } = await apiInstance.deleteUser();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**SuccessResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | User deleted successfully  |  -  |
|**0** | Error response  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getUser**
> UserResponse getUser()

Retrieves the profile and authorised apps of the authenticated user. Requires an authenticated user. 

### Example

```typescript
import {
    UsersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

const { status, data } = await apiInstance.getUser();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**UserResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | User profile retrieved successfully  |  -  |
|**0** | Error response  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

