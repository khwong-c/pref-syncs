# PreferencesApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**deletePref**](#deletepref) | **DELETE** /pref/{app} | Delete a preference |
|[**getPref**](#getpref) | **GET** /pref/{app} | Get a preference |
|[**postPref**](#postpref) | **POST** /pref/{app} | Update a preference |
|[**postPrefFromSource**](#postpreffromsource) | **POST** /pref/{app}/from/{src} | Update a preference from a source |
|[**startNotification**](#startnotification) | **GET** /pref/notification/{app} | Stream preference notifications |
|[**startNotificationFromSource**](#startnotificationfromsource) | **GET** /pref/notification/{app}/from/{src} | Stream preference notifications excluding a source |

# **deletePref**
> SuccessResponse deletePref()

Deletes the authenticated user\'s stored preference for the given app. Requires an authenticated user. 

### Example

```typescript
import {
    PreferencesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new PreferencesApi(configuration);

let app: string; //The unique identifier of the app  (default to undefined)

const { status, data } = await apiInstance.deletePref(
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
|**200** | Preference deleted successfully  |  -  |
|**0** | Error response  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getPref**
> PrefResponse getPref()

Retrieves the authenticated user\'s stored preference payload for the given app. Requires an authenticated user. 

### Example

```typescript
import {
    PreferencesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new PreferencesApi(configuration);

let app: string; //The unique identifier of the app  (default to undefined)

const { status, data } = await apiInstance.getPref(
    app
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **app** | [**string**] | The unique identifier of the app  | defaults to undefined|


### Return type

**PrefResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Preference retrieved successfully  |  -  |
|**0** | Error response  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **postPref**
> PrefResponse postPref(body)

Creates or updates the authenticated user\'s preference payload for the given app, subject to the app\'s payload size limit. Requires an authenticated user. 

### Example

```typescript
import {
    PreferencesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new PreferencesApi(configuration);

let app: string; //The unique identifier of the app  (default to undefined)
let body: any; //

const { status, data } = await apiInstance.postPref(
    app,
    body
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **body** | **any**|  | |
| **app** | [**string**] | The unique identifier of the app  | defaults to undefined|


### Return type

**PrefResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Preference updated successfully  |  -  |
|**0** | Error response  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **postPrefFromSource**
> PrefResponse postPrefFromSource(body)

Creates or updates the authenticated user\'s preference for the given app, tagging the update with a source identifier so it can be excluded from that source\'s own notification stream. Requires an authenticated user. 

### Example

```typescript
import {
    PreferencesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new PreferencesApi(configuration);

let app: string; //The unique identifier of the app  (default to undefined)
let src: string; //The source identifier  (default to undefined)
let body: any; //

const { status, data } = await apiInstance.postPrefFromSource(
    app,
    src,
    body
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **body** | **any**|  | |
| **app** | [**string**] | The unique identifier of the app  | defaults to undefined|
| **src** | [**string**] | The source identifier  | defaults to undefined|


### Return type

**PrefResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Preference updated successfully  |  -  |
|**0** | Error response  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **startNotification**
> Notification startNotification()

Opens a server-sent events stream that emits notifications whenever the authenticated user\'s preference for the given app changes. Requires an authenticated user. 

### Example

```typescript
import {
    PreferencesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new PreferencesApi(configuration);

let app: string; //The unique identifier of the app  (default to undefined)

const { status, data } = await apiInstance.startNotification(
    app
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **app** | [**string**] | The unique identifier of the app  | defaults to undefined|


### Return type

**Notification**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/event-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Server-sent events stream for preference notifications  |  -  |
|**0** | Error response  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **startNotificationFromSource**
> Notification startNotificationFromSource()

Opens a server-sent events stream of preference change notifications for the given app, excluding updates originating from the given source. Requires an authenticated user. 

### Example

```typescript
import {
    PreferencesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new PreferencesApi(configuration);

let app: string; //The unique identifier of the app  (default to undefined)
let src: string; //The source identifier  (default to undefined)

const { status, data } = await apiInstance.startNotificationFromSource(
    app,
    src
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **app** | [**string**] | The unique identifier of the app  | defaults to undefined|
| **src** | [**string**] | The source identifier  | defaults to undefined|


### Return type

**Notification**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/event-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Server-sent events stream for preference notifications  |  -  |
|**0** | Error response  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

