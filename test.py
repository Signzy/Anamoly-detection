import requests
import json



def send_request(testname,batch,stream):
	payload = {
		"stream": stream,
		"data": batch	
	}
	headers = {
	  'Content-Type': 'application/json'
	}

	# print(payload)
	r = requests.post("http://localhost:40404/sad/post",headers=headers, data=json.dumps(payload))

	resp = r.json()

	# print(resp)
	for key in resp["stats"]:
		if resp["stats"][key] == 1 :
			print("Anomaly detected",testname,key)


def run_tests(rl):
	for r in rl:
		send_request(r["testname"],r["batch"],r["stream"])






tlist = [
	{
		"testname":"t1",
		"batch":[
	        {
	            "name":"Ankur"
	        }
	    ],
	    "stream":"t1"
	},
	{
		"testname":"t1",
		"batch":[
	        {
	            "name":"Ankur"
	        }
	    ],
	    "stream":"t1"
	},
	{
		"testname":"t1",
		"batch":[
	        {
	            "name":"Ankur"
	        }
	    ],
	    "stream":"t1"
	},
	{
		"testname":"t1",
		"batch":[
	        {
	            "name":"Rajdeep Sharma"
	        }
	    ],
	    "stream":"t1"
	},
]





run_tests(tlist)









