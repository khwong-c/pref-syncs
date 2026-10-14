# GeneralApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getRoot**](#getroot) | **GET** / | Health check |

# **getRoot**
> string getRoot()

Returns a simple greeting to confirm the server is running. No authentication is required. 

### Example

```typescript
import {
    GeneralApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new GeneralApi(configuration);

const { status, data } = await apiInstance.getRoot();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/plain


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Server is running  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

