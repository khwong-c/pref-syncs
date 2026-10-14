# AppsApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createApp**](#createapp) | **POST** /app | Create an app |
|[**deleteApp**](#deleteapp) | **DELETE** /app/{app} | Delete an app |
|[**getApp**](#getapp) | **GET** /app/{app} | Get an app |
|[**modifyApp**](#modifyapp) | **PUT** /app/{app} | Modify an app |

# **createApp**
> AppResponse createApp(appRequest)

Creates a new app owned by the authenticated user. Requires an authenticated admin user. 

### Example

```typescript
import {
    AppsApi,
    Configuration,
    AppRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AppsApi(configuration);

let appRequest: AppRequest; //

const { status, data } = await apiInstance.createApp(
    appRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **appRequest** | **AppRequest**|  | |


### Return type

**AppResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Application created successfully  |  -  |
|**0** | Error response  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteApp**
> SuccessResponse deleteApp()

Permanently deletes an app by its ID. Requires an authenticated admin user. 

### Example

```typescript
import {
    AppsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AppsApi(configuration);

let app: string; //The unique identifier of the app  (default to undefined)

const { status, data } = await apiInstance.deleteApp(
    app
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **app** | [**string**] | The unique identifier of the app  | defaults to undefined|


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
|**200** | Application deleted successfully  |  -  |
|**0** | Error response  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getApp**
> AppResponse getApp()

Retrieves the details of an app by its ID. Requires an authenticated user. 

### Example

```typescript
import {
    AppsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AppsApi(configuration);

let app: string; //The unique identifier of the app  (default to undefined)

const { status, data } = await apiInstance.getApp(
    app
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **app** | [**string**] | The unique identifier of the app  | defaults to undefined|


### Return type

**AppResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Application retrieved successfully  |  -  |
|**0** | Error response  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **modifyApp**
> AppResponse modifyApp(appRequest)

Updates the name, description, or limits of an existing app. Requires an authenticated admin user. 

### Example

```typescript
import {
    AppsApi,
    Configuration,
    AppRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AppsApi(configuration);

let app: string; //The unique identifier of the app  (default to undefined)
let appRequest: AppRequest; //

const { status, data } = await apiInstance.modifyApp(
    app,
    appRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **appRequest** | **AppRequest**|  | |
| **app** | [**string**] | The unique identifier of the app  | defaults to undefined|


### Return type

**AppResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Application modified successfully  |  -  |
|**0** | Error response  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

