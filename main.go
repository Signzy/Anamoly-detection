package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"regexp"
	"github.com/google/uuid"
	"time"
	"strconv"
	"math"
	"sync"
)


const FEATURE_COUNT int = 5
const MAX_WINDOW_LENGTH int = 7


type SafeCounter struct {
	m sync.Mutex
	n int
}

func (c *SafeCounter) Get() int {
	c.m.Lock()
	defer c.m.Unlock()

	c.n += 1
	if c.n >= MAX_WINDOW_LENGTH {
		c.n = 0
	}
	return c.n
}

func (c *SafeCounter) Set(delta int) {
	c.m.Lock()
	c.n = delta
	c.m.Unlock()
}

type AD_Block struct{
	Stream string
	ID string 
	Key string 
	Timestamp string 
	Features [FEATURE_COUNT]float64
}

type Stats_Block struct{
	Avg [FEATURE_COUNT]float64
	Std [FEATURE_COUNT]float64
}


type Window_Block struct{
	W_avg float64
	W_std float64
	Window [MAX_WINDOW_LENGTH]*AD_Block
	W_write_location *SafeCounter
	Total_writes uint64
}


type Single_Key_Prediction struct {
	Key string `json:"key"`
	Prediction int `json:"prediction"`
}

type Single_Block_Prediction struct {
	ID string `json:"id"` 
	// Timestamp string
	Predictions []Single_Key_Prediction `json:"predictions"`
}

var G_stats = make(map[string]*Window_Block)


func get_int_features(s string) ([FEATURE_COUNT]float64) {

	var feature_array [FEATURE_COUNT]float64 

	for i := 0; i < FEATURE_COUNT; i++ {
		feature_array[i] = 0.0
	}

	f0, _ := strconv.ParseFloat(s,64)

	feature_array[0] = f0

	return feature_array
}

func get_str_features(s string) (feature_array [FEATURE_COUNT]float64){

	/*
	features for strings
		f1 = string length
		f2 = number of alphabets
		f3 = number of numbers [0..9]
		f4 = number of spaces
		f5 = number of non - alphabet/numeric/space characters
	*/

	f1 := float64(len(s))
	f2 := 0.0
	f3 := 0.0
	f4 := 0.0
	f5 := 0.0

	for i := range s {

		ascii_value := int(rune(s[i]))

		if (ascii_value >= 65 && ascii_value <= 90) || 
			(ascii_value >= 97 && ascii_value <= 122) {
				f2++
		} else if (ascii_value >= 48 && ascii_value <= 57){
				f3++
		} else if (ascii_value == 32){
				f4++
		} else {
				f5++
		}
	}

	feature_array = [FEATURE_COUNT]float64{f1,f2,f3,f4,f5}
	return feature_array
}


func calculate_mean(arr [MAX_WINDOW_LENGTH]float64, count int) float64{
	
	var sum float64 = 0.0
	for i := 0; i < count; i++ {
		sum += arr[i]
	}
	var avg float64 = sum / float64(count)
	return avg
}


func calculate_stddev(arr [MAX_WINDOW_LENGTH]float64, count int, mean float64) float64 {
	
	var sd float64
	for j := 0; j < count; j++ {
        sd += math.Pow(arr[j] - mean, 2)
    }
    sd = math.Sqrt(sd/float64(count))
    return sd
}

func get_window_stats(queue [MAX_WINDOW_LENGTH]*AD_Block) Stats_Block {

	/*
		create slice for each feature
		pass each slice to calculate average function
		store all averages in a new slice
	*/

	copy_of_window := [FEATURE_COUNT][MAX_WINDOW_LENGTH]float64{}

	var avg [FEATURE_COUNT]float64
	var std [FEATURE_COUNT]float64

	var count int = 0

	for count < MAX_WINDOW_LENGTH {
    	
    	ad_block := queue[count]

		for i := 0; i < FEATURE_COUNT; i++ {
			copy_of_window[i][count] = ad_block.Features[i]
		}	
		count++	 
	}


	for i := 0; i < FEATURE_COUNT; i++ {
		avg[i] = calculate_mean(copy_of_window[i],count)
		std[i] = calculate_stddev(copy_of_window[i],count,avg[i])
	}

	stats := Stats_Block{Avg : avg, Std : std}

	return stats
}

func predict_anomaly(ad_block *AD_Block, window_stats_block Stats_Block) int {

	for i := 0; i < FEATURE_COUNT; i++ {
		diff := math.Abs(ad_block.Features[i] - window_stats_block.Avg[i])
		mult := 2.0
		if diff > mult*window_stats_block.Std[i] {
			return 1
		}
	}
	return 0
}


func process(c *gin.Context) {

	/*
		input 
		{
			"stream": "streamname",
			"data":[
				{
					"name":"rajdeep",
					"dob":"20-10-1997",
					"pan":"ABCD12345F"
				}
			]
		}

		add an id for each data entry
		add timestamp
		for key value pair in a data object
			append stream + id + key_name 
			if it is a string extract the features f1 to f4
			if it an int leave it in as f1 and keep f2..4 empty
		save stream + key to get global features

		[
			{
				"stream": "stream",
				"id": "id",
				"key":"keyname",
				"timestamp":"ts",
				"f1":1,
				"f2":0,
				"f3":0,
				"f4":0
			}
		]

		use stream+keyname to get global stats



	*/

	var request_object map[string]*json.RawMessage

	jsonData, err := c.GetRawData() // []byte

	if err != nil {
		fmt.Println(err)
	    c.JSON(http.StatusOK, gin.H{
		  "message": "error",
		})
		return
	}

	if err := json.Unmarshal(jsonData, &request_object); err != nil {
    	fmt.Println(err)
    	c.JSON(http.StatusOK, gin.H{
		  "message": "error",
		})
		return
	}

	stream_str := string(*request_object["stream"])
	stream_str = stream_str[1:len(stream_str)-1]

	fmt.Println("Stream: ",stream_str)
	
	var data []*json.RawMessage
	
	if err := json.Unmarshal(*request_object["data"], &data); err != nil {
    	fmt.Println(err)
    	c.JSON(http.StatusOK, gin.H{	
		  "message": "error",
		})
		return
	}


	var predictions []Single_Block_Prediction


	for i := range data {

		var prediction int = -1
		id := uuid.New().String()

		fmt.Println("Block Values:",id)
		
		var block map[string]*json.RawMessage

		if err := json.Unmarshal(*data[i], &block); err != nil {
	    	fmt.Println(err)
	    	c.JSON(http.StatusOK, gin.H{
			  "message": "error",
			})
			return
		}
	    
		var key_predictions []Single_Key_Prediction    

	    for key,value := range block {

	    	
    		timestamp := time.Now().Format(time.RFC850)
	    	key_str := string(key)

	    	isString,_ := regexp.Match(`".*"`,*value)
	    	var value_str string = string(*value)
	    	var ad_block *AD_Block

	    	if isString {
	    		
	    		feature_array := get_str_features(value_str[1:len(value_str)-1])
	    		fmt.Println(key_str,value_str[1:len(value_str)-1],feature_array)
	    		
	    		ad_block = &AD_Block{
	    			Stream:stream_str, 
	    			ID:id,
	    			Key:key_str,
	    			Timestamp:timestamp,
	    			Features:feature_array,
	    		}
	    	} else {
	    		
				feature_array := get_int_features(value_str)

	    		//  int or float value
	    		ad_block = &AD_Block{
	    			Stream:stream_str, 
	    			ID:id,
	    			Key:key_str,
	    			Timestamp:timestamp,
	    			Features:feature_array,
	    		}
	    	}

	    	var window_slug string = stream_str+"#"+key_str

	    	if _, ok := G_stats[window_slug]; ok {
				
				current_loc := G_stats[window_slug].W_write_location.Get()	
				G_stats[window_slug].Window[current_loc] = ad_block

			} else {
				var window_array [MAX_WINDOW_LENGTH]*AD_Block
				var window_write_location_counter = &SafeCounter{}
				window_write_location_counter.Set(0)

		    	G_stats[window_slug] = &Window_Block{
		    		W_avg:0,
		    		W_std:0,
		    		Window: window_array,
		    		W_write_location: window_write_location_counter,
		    		Total_writes:0,
		    	}
				G_stats[window_slug].Window[0] = ad_block
		    }

		    total_predictions := G_stats[window_slug].Total_writes
			if total_predictions > uint64(MAX_WINDOW_LENGTH) {
				window_stats_block := get_window_stats(G_stats[window_slug].Window)
				prediction = predict_anomaly(ad_block,window_stats_block)
			}

			key_prediction_struct := Single_Key_Prediction{key_str,prediction}

			key_predictions = append(key_predictions,key_prediction_struct)

			G_stats[window_slug].Total_writes += 1

	    } //key-value loop

	    block_pred := Single_Block_Prediction{id,key_predictions}

	    predictions = append(predictions,block_pred)
	
	} // data blocks loop

	// fmt.Printf("%+v\n",*G_stats["pan_extraction#dob"].Window[0])

	c.JSON(http.StatusOK, gin.H{
	  "message": "pong",
	  "result": predictions,
	})
}

func main(){

	fmt.Println("Anomaly Detection");

	r := gin.Default()
	r.POST("/sad/post", process)
	r.Run()

}
