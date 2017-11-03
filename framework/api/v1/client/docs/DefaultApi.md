# \DefaultApi

All URIs are relative to *http://localhost/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddCollector**](DefaultApi.md#AddCollector) | **Post** /collectors | 
[**DeleteCollector**](DefaultApi.md#DeleteCollector) | **Delete** /collectors/{collector-name} | 
[**GetCollector**](DefaultApi.md#GetCollector) | **Get** /collectors/{collector-name} | 
[**GetCollectors**](DefaultApi.md#GetCollectors) | **Get** /collectors | 
[**GetHealthz**](DefaultApi.md#GetHealthz) | **Get** /healthz | 
[**GetLiveness**](DefaultApi.md#GetLiveness) | **Get** /liveness | 
[**GetMetrics**](DefaultApi.md#GetMetrics) | **Get** /metrics | 
[**HeartbeatCollector**](DefaultApi.md#HeartbeatCollector) | **Get** /collectors/{collector-name}/heartbeat | 
[**RemoveAllFields**](DefaultApi.md#RemoveAllFields) | **Delete** /edges/{edge-type}/names/{field-name} | 
[**RemoveField**](DefaultApi.md#RemoveField) | **Delete** /edges/{edge-type}/key/{edge-key}/names/{field-name} | 
[**UpdateCollector**](DefaultApi.md#UpdateCollector) | **Post** /collectors/{collector-name} | 
[**UpsertField**](DefaultApi.md#UpsertField) | **Post** /edges/{edge-type}/names/{field-name} | 


# **AddCollector**
> AddCollector($body)



inserts new collector


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

# **DeleteCollector**
> DeleteCollector($collectorName)



delete Collector by Name


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **collectorName** | **string**|  | 

### Return type

void (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetCollector**
> Collector GetCollector($collectorName)



get collector by name


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **collectorName** | **string**|  | 

### Return type

[**Collector**](Collector.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetCollectors**
> []Collector GetCollectors()



get all collector services


### Parameters
This endpoint does not need any parameter.

### Return type

[**[]Collector**](Collector.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetHealthz**
> GetHealthz()



get health of framework


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

# **GetLiveness**
> GetLiveness()



get liveness of framework


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

# **GetMetrics**
> GetMetrics()



lets prometheus scrape the framework


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

# **HeartbeatCollector**
> Collector HeartbeatCollector($collectorName)



heartbeat endpoint for a specific collector


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **collectorName** | **string**|  | 

### Return type

[**Collector**](Collector.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RemoveAllFields**
> RemoveAllFields($edgeType, $fieldName)



endpoint to remove all fieldNames from all edges


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **edgeType** | **string**|  | 
 **fieldName** | **string**|  | 

### Return type

void (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RemoveField**
> RemoveField($edgeType, $edgeKey, $fieldName)



endpoint to remove a value from an edge


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **edgeType** | **string**|  | 
 **edgeKey** | **string**|  | 
 **fieldName** | **string**|  | 

### Return type

void (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateCollector**
> UpdateCollector($collectorName, $body)



updates collector


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **collectorName** | **string**|  | 
 **body** | [**Collector**](Collector.md)| Collector object | 

### Return type

void (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpsertField**
> UpsertField($edgeType, $fieldName, $body)



endpoint for collectors to add services


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **edgeType** | **string**|  | 
 **fieldName** | **string**|  | 
 **body** | [**EdgeScore**](EdgeScore.md)| Value object | 

### Return type

void (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

