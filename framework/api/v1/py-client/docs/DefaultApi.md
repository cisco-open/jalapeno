# client.DefaultApi

All URIs are relative to *http://localhost/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**add_collector**](DefaultApi.md#add_collector) | **POST** /collectors | 
[**delete_collector**](DefaultApi.md#delete_collector) | **DELETE** /collectors/{collector-name} | 
[**get_collector**](DefaultApi.md#get_collector) | **GET** /collectors/{collector-name} | 
[**get_collectors**](DefaultApi.md#get_collectors) | **GET** /collectors | 
[**get_healthz**](DefaultApi.md#get_healthz) | **GET** /healthz | 
[**get_liveness**](DefaultApi.md#get_liveness) | **GET** /liveness | 
[**get_metrics**](DefaultApi.md#get_metrics) | **GET** /metrics | 
[**heartbeat_collector**](DefaultApi.md#heartbeat_collector) | **GET** /collectors/{collector-name}/heartbeat | 
[**update_collector**](DefaultApi.md#update_collector) | **POST** /collectors/{collector-name} | 


# **add_collector**
> add_collector(body)



inserts new collector

### Example 
```python
from __future__ import print_statement
import time
import client
from client.rest import ApiException
from pprint import pprint

# Configure API key authorization: Bearer
client.configuration.api_key['Authorization'] = 'YOUR_API_KEY'
# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# client.configuration.api_key_prefix['Authorization'] = 'Bearer'

# create an instance of the API class
api_instance = client.DefaultApi()
body = client.Collector() # Collector | Collector object

try: 
    api_instance.add_collector(body)
except ApiException as e:
    print("Exception when calling DefaultApi->add_collector: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**Collector**](Collector.md)| Collector object | 

### Return type

void (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **delete_collector**
> delete_collector(collector_name)



delete Collector by Name

### Example 
```python
from __future__ import print_statement
import time
import client
from client.rest import ApiException
from pprint import pprint

# Configure API key authorization: Bearer
client.configuration.api_key['Authorization'] = 'YOUR_API_KEY'
# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# client.configuration.api_key_prefix['Authorization'] = 'Bearer'

# create an instance of the API class
api_instance = client.DefaultApi()
collector_name = 'collector_name_example' # str | 

try: 
    api_instance.delete_collector(collector_name)
except ApiException as e:
    print("Exception when calling DefaultApi->delete_collector: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **collector_name** | **str**|  | 

### Return type

void (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_collector**
> Collector get_collector(collector_name)



get collector by name

### Example 
```python
from __future__ import print_statement
import time
import client
from client.rest import ApiException
from pprint import pprint

# Configure API key authorization: Bearer
client.configuration.api_key['Authorization'] = 'YOUR_API_KEY'
# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# client.configuration.api_key_prefix['Authorization'] = 'Bearer'

# create an instance of the API class
api_instance = client.DefaultApi()
collector_name = 'collector_name_example' # str | 

try: 
    api_response = api_instance.get_collector(collector_name)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling DefaultApi->get_collector: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **collector_name** | **str**|  | 

### Return type

[**Collector**](Collector.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_collectors**
> list[Collector] get_collectors()



get all collector services

### Example 
```python
from __future__ import print_statement
import time
import client
from client.rest import ApiException
from pprint import pprint

# Configure API key authorization: Bearer
client.configuration.api_key['Authorization'] = 'YOUR_API_KEY'
# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# client.configuration.api_key_prefix['Authorization'] = 'Bearer'

# create an instance of the API class
api_instance = client.DefaultApi()

try: 
    api_response = api_instance.get_collectors()
    pprint(api_response)
except ApiException as e:
    print("Exception when calling DefaultApi->get_collectors: %s\n" % e)
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**list[Collector]**](Collector.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_healthz**
> get_healthz()



get health of framework

### Example 
```python
from __future__ import print_statement
import time
import client
from client.rest import ApiException
from pprint import pprint

# Configure API key authorization: Bearer
client.configuration.api_key['Authorization'] = 'YOUR_API_KEY'
# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# client.configuration.api_key_prefix['Authorization'] = 'Bearer'

# create an instance of the API class
api_instance = client.DefaultApi()

try: 
    api_instance.get_healthz()
except ApiException as e:
    print("Exception when calling DefaultApi->get_healthz: %s\n" % e)
```

### Parameters
This endpoint does not need any parameter.

### Return type

void (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_liveness**
> get_liveness()



get liveness of framework

### Example 
```python
from __future__ import print_statement
import time
import client
from client.rest import ApiException
from pprint import pprint

# Configure API key authorization: Bearer
client.configuration.api_key['Authorization'] = 'YOUR_API_KEY'
# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# client.configuration.api_key_prefix['Authorization'] = 'Bearer'

# create an instance of the API class
api_instance = client.DefaultApi()

try: 
    api_instance.get_liveness()
except ApiException as e:
    print("Exception when calling DefaultApi->get_liveness: %s\n" % e)
```

### Parameters
This endpoint does not need any parameter.

### Return type

void (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_metrics**
> get_metrics()



lets prometheus scrape the framework

### Example 
```python
from __future__ import print_statement
import time
import client
from client.rest import ApiException
from pprint import pprint

# Configure API key authorization: Bearer
client.configuration.api_key['Authorization'] = 'YOUR_API_KEY'
# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# client.configuration.api_key_prefix['Authorization'] = 'Bearer'

# create an instance of the API class
api_instance = client.DefaultApi()

try: 
    api_instance.get_metrics()
except ApiException as e:
    print("Exception when calling DefaultApi->get_metrics: %s\n" % e)
```

### Parameters
This endpoint does not need any parameter.

### Return type

void (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **heartbeat_collector**
> Collector heartbeat_collector(collector_name)



heartbeat endpoint for a specific collector

### Example 
```python
from __future__ import print_statement
import time
import client
from client.rest import ApiException
from pprint import pprint

# Configure API key authorization: Bearer
client.configuration.api_key['Authorization'] = 'YOUR_API_KEY'
# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# client.configuration.api_key_prefix['Authorization'] = 'Bearer'

# create an instance of the API class
api_instance = client.DefaultApi()
collector_name = 'collector_name_example' # str | 

try: 
    api_response = api_instance.heartbeat_collector(collector_name)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling DefaultApi->heartbeat_collector: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **collector_name** | **str**|  | 

### Return type

[**Collector**](Collector.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **update_collector**
> update_collector(collector_name, body)



updates collector

### Example 
```python
from __future__ import print_statement
import time
import client
from client.rest import ApiException
from pprint import pprint

# Configure API key authorization: Bearer
client.configuration.api_key['Authorization'] = 'YOUR_API_KEY'
# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# client.configuration.api_key_prefix['Authorization'] = 'Bearer'

# create an instance of the API class
api_instance = client.DefaultApi()
collector_name = 'collector_name_example' # str | 
body = client.Collector() # Collector | Collector object

try: 
    api_instance.update_collector(collector_name, body)
except ApiException as e:
    print("Exception when calling DefaultApi->update_collector: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **collector_name** | **str**|  | 
 **body** | [**Collector**](Collector.md)| Collector object | 

### Return type

void (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

