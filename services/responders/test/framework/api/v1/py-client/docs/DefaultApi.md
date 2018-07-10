# client.DefaultApi

All URIs are relative to *http://localhost/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**add_collector**](DefaultApi.md#add_collector) | **POST** /collectors | 
[**delete_collector**](DefaultApi.md#delete_collector) | **DELETE** /collectors/{collector-name} | 
[**get_collector**](DefaultApi.md#get_collector) | **GET** /collectors/{collector-name} | 
[**get_collectors**](DefaultApi.md#get_collectors) | **GET** /collectors | 
[**get_edge**](DefaultApi.md#get_edge) | **GET** /edges/{edge-type}/filter/{field-name}/{field-value} | 
[**get_healthz**](DefaultApi.md#get_healthz) | **GET** /healthz | 
[**get_liveness**](DefaultApi.md#get_liveness) | **GET** /liveness | 
[**get_metrics**](DefaultApi.md#get_metrics) | **GET** /metrics | 
[**heartbeat_collector**](DefaultApi.md#heartbeat_collector) | **GET** /collectors/{collector-name}/heartbeat | 
[**query_arango**](DefaultApi.md#query_arango) | **GET** /query/{Collection} | 
[**remove_all_fields**](DefaultApi.md#remove_all_fields) | **DELETE** /edges/{edge-type}/names/{field-name} | 
[**remove_field**](DefaultApi.md#remove_field) | **DELETE** /edges/{edge-type}/key/{edge-key}/names/{field-name} | 
[**update_collector**](DefaultApi.md#update_collector) | **POST** /collectors/{collector-name} | 
[**upsert_field**](DefaultApi.md#upsert_field) | **PUT** /edges/{edge-type}/names/{field-name} | 


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
> list[Collector] get_collectors(name=name, description=description, status=status, edge_type=edge_type, field_name=field_name, timeout=timeout, last_heartbeat=last_heartbeat)



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
name = 'name_example' # str |  (optional)
description = 'description_example' # str |  (optional)
status = 'status_example' # str |  (optional)
edge_type = 'edge_type_example' # str |  (optional)
field_name = 'field_name_example' # str |  (optional)
timeout = 'timeout_example' # str |  (optional)
last_heartbeat = 'last_heartbeat_example' # str |  (optional)

try: 
    api_response = api_instance.get_collectors(name=name, description=description, status=status, edge_type=edge_type, field_name=field_name, timeout=timeout, last_heartbeat=last_heartbeat)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling DefaultApi->get_collectors: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **str**|  | [optional] 
 **description** | **str**|  | [optional] 
 **status** | **str**|  | [optional] 
 **edge_type** | **str**|  | [optional] 
 **field_name** | **str**|  | [optional] 
 **timeout** | **str**|  | [optional] 
 **last_heartbeat** | **str**|  | [optional] 

### Return type

[**list[Collector]**](Collector.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_edge**
> object get_edge(edge_type, field_name, field_value)



get edge with field name/value

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
edge_type = 'edge_type_example' # str | 
field_name = 'field_name_example' # str | 
field_value = 'field_value_example' # str | 

try: 
    api_response = api_instance.get_edge(edge_type, field_name, field_value)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling DefaultApi->get_edge: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **edge_type** | **str**|  | 
 **field_name** | **str**|  | 
 **field_value** | **str**|  | 

### Return type

**object**

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

# **query_arango**
> object query_arango(collection)



query arango for edges and nodes

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
collection = 'collection_example' # str | 

try: 
    api_response = api_instance.query_arango(collection)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling DefaultApi->query_arango: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **collection** | **str**|  | 

### Return type

**object**

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **remove_all_fields**
> remove_all_fields(edge_type, field_name)



endpoint to remove all fieldNames from all edges

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
edge_type = 'edge_type_example' # str | 
field_name = 'field_name_example' # str | 

try: 
    api_instance.remove_all_fields(edge_type, field_name)
except ApiException as e:
    print("Exception when calling DefaultApi->remove_all_fields: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **edge_type** | **str**|  | 
 **field_name** | **str**|  | 

### Return type

void (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **remove_field**
> remove_field(edge_type, edge_key, field_name)



endpoint to remove a value from an edge

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
edge_type = 'edge_type_example' # str | 
edge_key = 'edge_key_example' # str | 
field_name = 'field_name_example' # str | 

try: 
    api_instance.remove_field(edge_type, edge_key, field_name)
except ApiException as e:
    print("Exception when calling DefaultApi->remove_field: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **edge_type** | **str**|  | 
 **edge_key** | **str**|  | 
 **field_name** | **str**|  | 

### Return type

void (empty response body)

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

# **upsert_field**
> upsert_field(edge_type, field_name, body)



endpoint for collectors to add services

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
edge_type = 'edge_type_example' # str | 
field_name = 'field_name_example' # str | 
body = client.EdgeScore() # EdgeScore | Value object

try: 
    api_instance.upsert_field(edge_type, field_name, body)
except ApiException as e:
    print("Exception when calling DefaultApi->upsert_field: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **edge_type** | **str**|  | 
 **field_name** | **str**|  | 
 **body** | [**EdgeScore**](EdgeScore.md)| Value object | 

### Return type

void (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

