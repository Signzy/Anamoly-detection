Listing down scenarios that need to be passed through simulation into the anomaly detection engine.




# Pipeline examples

Below JSON is as of now for representative purpose only. This can be updated as per agent that's sending to the transformer.

Can have concept of index (<suffix>) like elasticsearch.

```json

	{
		"source": "<source>",
		"data": [{
			"name": "Ankur Pandey",
			"dob": "18/06/1992",
			"fatherName": "Sanjay Pandey",
			"pan": "Rajdeep S Sharma",
		}]
	}

```

The hit recieved by transformer above gets appended with a timestamp value of the current time as on the reciever, before further sending it to the core.

```json

	[{
		"timestamp": "...,",
		"data": [{
			"stream": "<suffix>_name",
			"value": "Ankur Pandey"
		}, {
			"stream": "<suffix>_dob",
			"value": "Ankur Pandey"
		}, {
			"stream": "<suffix>_fatherName",
			"value": "Ankur Pandey"
		}, {
			"stream": "<suffix>_pan",
			"value": "Ankur Pandey"
		}]
	}]

```

Such data comes 1000 times, after that for 20 times, it sends below data.


```json

	{
		"name1": "",
		"dob": "Zain Ahmed",
		"name3": "Vedant",
		"name4": "Rajdeep S Sharma",
	}

```




# AD inputs

Below is an example of string input.

```js
	// This is sending to a particular <suffix>_name as defined below.
	[
		{timestamp: "<current-timestamp>", "value":  "Ankur Pandey"},
		{timestamp: "<current-timestamp>", "value":  "Zain Ahmed"},
		{timestamp: "<current-timestamp>", "value":  "Vedant"},
		{timestamp: "<current-timestamp>", "value":  "Rajdeep S Sharma"},
		"....."
	]

```


```js

	[
		{timestamp: "<current-timestamp>",  "Ankur Pandey"},
		{timestamp: "<current-timestamp>",  "Zain Ahmed"},
		{timestamp: "<current-timestamp>",  "Vedant"},
		{timestamp: "<current-timestamp>",  "Rajdeep S Sharma"},
		"....."
	]

```



# Math model input

