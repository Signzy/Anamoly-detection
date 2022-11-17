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
const MAX_WINDOW_LENGTH int = 5000

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


func calculate_mean(arr []float64, count int) float64{
	
	var sum float64 = 0.0
	for i := 0; i < count; i++ {
		sum += arr[i]
	}
	var avg float64 = sum / float64(count)
	return avg
}


func calculate_stddev(arr []float64, count int, mean float64) float64 {
	
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
		avg[i] = calculate_mean(copy_of_window[i][:],count)
		std[i] = calculate_stddev(copy_of_window[i][:],count,avg[i])
	}

	stats := Stats_Block{Avg : avg, Std : std}

	return stats
}


func get_batch_stats(queue []*AD_Block) Stats_Block {

	copy_of_window := [FEATURE_COUNT][]float64{}

	var avg [FEATURE_COUNT]float64
	var std [FEATURE_COUNT]float64

	var count int = 0

	for count < len(queue) {
    	
    	ad_block := queue[count]

		for i := 0; i < FEATURE_COUNT; i++ {
			copy_of_window[i] = append(copy_of_window[i],ad_block.Features[i])
		}	
		count++	 
	}

	for i := 0; i < FEATURE_COUNT; i++ {
		avg[i] = calculate_mean(copy_of_window[i][:],len(queue))
		std[i] = calculate_stddev(copy_of_window[i][:],len(queue),avg[i])
	}

	stats := Stats_Block{Avg : avg, Std : std}

	return stats
}

func predict_anomaly(batch_stats_block Stats_Block, window_stats_block Stats_Block) int {

	for i := 0; i < FEATURE_COUNT; i++ {
		diff := math.Abs(batch_stats_block.Avg[i] - window_stats_block.Avg[i])
		mult := 2.0
		if diff > mult*window_stats_block.Std[i] {
			return 1
		}
	}
	return 0
}


func reset(c *gin.Context) {

	var request_object map[string]string 
	jsonData, err := c.GetRawData()
	if err != nil {
		fmt.Println(err)
	    c.JSON(http.StatusOK, gin.H{
		  "message": "error4",
		})
		return
	}

	if err := json.Unmarshal(jsonData, &request_object); err != nil {
    	fmt.Println(err)
    	c.JSON(http.StatusOK, gin.H{
		  "message": "error0",
		})
		return
	}

	window_slug := request_object["buffer"]

	if _, ok := G_stats[window_slug]; ok {
		G_stats[window_slug].W_write_location.Set(0)
		G_stats[window_slug].Total_writes = 0
	
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			  "message": "no such buffer exists",
		})
		return	
	}

	

	c.JSON(http.StatusOK, gin.H{
		  "message": "buffer has been reset",
	})
	return
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
		  "message": "error4",
		})
		return
	}

	if err := json.Unmarshal(jsonData, &request_object); err != nil {
    	fmt.Println(err)
    	c.JSON(http.StatusOK, gin.H{
		  "message": "error0",
		})
		return
	}

	stream_str := string(*request_object["stream"])
	stream_str = stream_str[1:len(stream_str)-1]

	// fmt.Println("Stream: ",stream_str)
	
	var data []*json.RawMessage
	
	if err := json.Unmarshal(*request_object["data"], &data); err != nil {
    	fmt.Println(err)
    	c.JSON(http.StatusOK, gin.H{	
		  "message": "error1",
		})
		return
	}

	batch_stats := make(map[string]int)

	batch_block_store := make(map[string][]*AD_Block)

	var prediction int = -1
	
	timestamp := time.Now().Format(time.RFC850)

	for i := range data {

		id := uuid.New().String()
		

		// fmt.Println("Block Values:",id)
		
		var block map[string]*json.RawMessage

		if err := json.Unmarshal(*data[i], &block); err != nil {
	    	fmt.Println(err)
	    	c.JSON(http.StatusOK, gin.H{
			  "message": "error2",
			})
			return
		}
	    
	    for key,value := range block {

	    	// todo: check for int and float 

	    	key_str := string(key)

	    	isString,_ := regexp.Match(`".*"`,*value)
	    	var value_str string = string(*value)
	    	var ad_block *AD_Block

	    	if isString {
	    		
	    		feature_array := get_str_features(value_str[1:len(value_str)-1])
	    		// fmt.Println(key_str,value_str[1:len(value_str)-1],feature_array)
	    		
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
			batch_block_store[window_slug] = append(batch_block_store[window_slug],ad_block)

	    } //key-value loop

	} // data blocks loop

    for window_slug,ad_block_slice := range batch_block_store {

    	if _, ok := G_stats[window_slug]; ok {
			
			total_predictions := G_stats[window_slug].Total_writes
			if total_predictions >= uint64(MAX_WINDOW_LENGTH) {
				window_stats_block := get_window_stats(G_stats[window_slug].Window)
				batch_stats_block := get_batch_stats(ad_block_slice)
				// fmt.Printf("%+v\n",window_slug)
				// fmt.Printf("%+v\n",batch_stats_block)
				// fmt.Printf("%+v\n",window_stats_block)
				prediction = predict_anomaly(batch_stats_block,window_stats_block)
			}

		}

		batch_stats[window_slug] = prediction
    
    } // make predctions loop

    for window_slug,ad_block_slice := range batch_block_store {

    	for i := range ad_block_slice {

    		ad_block := ad_block_slice[i]

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

		    G_stats[window_slug].Total_writes += 1
		}

		// fmt.Printf("%+v\n",G_stats[window_slug])

    } // over-write window loop

	// fmt.Printf("%+v\n",batch_stats)

	c.JSON(http.StatusOK, gin.H{
	  "message": "success",
	  "stats":batch_stats,
	})
}

func main(){

	fmt.Println("Anomaly Detection");

	r := gin.Default()
	r.POST("/sad/post", process)
	r.POST("/sad/reset", reset)
	r.Run("0.0.0.0:40404")

}
