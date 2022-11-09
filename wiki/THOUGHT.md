# Anomaly detection

We are creating this system in order to identify anomalies as they happen in a live environemnt. It is important to immediately catch failures in a complex uncertain environment where the complexity inhibits detection and correlation. 

Core philosophy is achieving observability without the need to setup / configure alerts and alert rules manually. You just pipe the data stream you want to monitor and the system will monitor it continuosly for you.

## Expectations 

An overall system is composed of hudreds of microserviecs and external services, which are in turn dependent on each other. A failure on one of the components or services can cascade to a more complex failure at a different place.

# Sample Scenarios

## Scenario 1:

API 1 one dependent upon API 2. API 2 is responsible for fetching cached data from a DB like MongoDB. 

### Failure:

Another system using the same DB instance is hitting slow queries impacting the DB performance.

### Impact:

API 1 response time is suddenly increased due to slowness in DB which was used by API2 which is now causing performance impact.

### Expecation from AD:

The system should detect anomaly in
* response time of the API 1
* response time of API 2, API 2 
* Spike in slow queries in MongoDB. 

With all three alerts triggering at the same time, it can be immediately deduced that there is a *connection between slowness of external facing API 1 and slowness of MongDB*, and to resolve the API 1 slowness, you can focus on slowness of the DB.

## Scenario 2:

API 1 provides data collecting it from API 2 and 3 and combines them to make a complete response. API 1 has fail all mechanism where any field if missing from the dependency APIs 2 and 3 is missing, then it passes a blank string in that field.

### Success scenario
```json

	{
		"field_1_from_api_2": "field_value_1",
		"field_2_from_api_2": "field_value_2",
		"field_3_from_api_2": "field_value_3",

		"field_1_from_api_3": "field_value_1",
		"field_2_from_api_3": "field_value_2",
		"field_3_from_api_3": "field_value_3",
	}

```

### Failed scenario

Given API 3, which is external to us is failing, we are catching the failure and giving blank response. 

```json

	{
		"field_1_from_api_2": "field_value_1",
		"field_2_from_api_2": "field_value_2",
		"field_3_from_api_2": "field_value_3",

		"field_1_from_api_3": "",
		"field_2_from_api_3": "",
		"field_3_from_api_3": "",
	}

```

### Expectation from AD:

Detect 
* that the certain fields in the API 1 have started to give blank response.
* API 2 is giving error status codes.

## Scenario 3

An HTTP endpoint which was fine with status codes was continuosly failing for few min with non 200 status codes for 3 min before falling back to normal behaviour. 

### Expectation from AD:
Detect the failures in the API and trigger relevant anomaly alerts.




## Data sources

Data can be taken from different sources as live streams, some common examples can be data provided as input JSON, HTTP response times & status codes. 




