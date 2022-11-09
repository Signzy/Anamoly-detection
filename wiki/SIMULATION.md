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
		"id":123,
		"data": [{
			"stream": "<suffix>_name",
			"value": "Ankur Pandey"
		}, {
			"stream": "<suffix>_dob",
			"value": "20-10-1997"
		}, {
			"stream": "<suffix>_fatherName",
			"value": "Father Pandey"
		}, {
			"stream": "<suffix>_pan",
			"value": "JHASB12312B"
		}]
	}]
```

Such data comes 1000 times, after that for 20 times, it sends below data.


```json

	[{
		"timestamp": "...,",
		"id":456,
		"data": [{
			"stream": "<suffix>_name",
			"value": " "
		}, {
			"stream": "<suffix>_dob",
			"value": "ASDASD"
		}, {
			"stream": "<suffix>_fatherName",
			"value": "********************************"
		}, {
			"stream": "<suffix>_pan",
			"value": "Q"
		}]
	}]
```



# AD inputs

Below is an example of string input.

```js
	{
		"stream":"stream_name",
		"series":[
			{id: "123", "value_str":  "Ankur Pandey",  "type":"string"},
			{id: "456", "value_str":  "Zain Ahmed",    "type":"string"},
			{id: "789", "value_str":  "Vedant", 	   "type":"string"},
			{id: "901", "value_str":  "Rajdeep Sharma","type":"string"},
			"....."
		]
	}

```


```js
	{
		"stream":"stream_name",
		"series":[
			{id: "123", "value_int":  123,  "type":"int"},
			{id: "456", "value_int":  123,  "type":"int"},
			{id: "789", "value_int":  345, 	"type":"int"},
			{id: "901", "value_int":  67,   "type":"int"},
			"....."
		]
	}

```


# Math model input

```js
	{
		"stream":"stream_name",
		"series":[
			{id: "123", "f1":  123,  "f2":123, "f3":876},
			{id: "456", "f1":  123,  "f2":123, "f3":876},
			{id: "789", "f1":  345,  "f2":123, "f3":876},
			{id: "901", "f1":  67,   "f2":123, "f3":876},
			"....."
		]
	}

```









