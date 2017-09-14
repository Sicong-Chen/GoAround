package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"log"
	"strconv"
)

const (
	DISTANCE = "200km"
)

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Post struct {	// struct: likewise, Java class
	// `json:"user"` is for the json parsing of this User field. Otherwise, by default it's 'User'.
	User     string `json:"user"`	// name + type + json parsing of this User field
	Message  string  `json:"message"`
	Location Location `json:"location"`
}

func main() {
	fmt.Println("started-service")
	http.HandleFunc("/post", handlerPost)		// url mapping
	http.HandleFunc("/search", handlerSearch)     // add an url mapping
	log.Fatal(http.ListenAndServe(":8080", nil))	// Port monitor
}


func handlerPost(w http.ResponseWriter, r *http.Request) {	// deal with POST operation
	// Parse from body of request to get a json object.
	fmt.Println("Received one post request")
	decoder := json.NewDecoder(r.Body)
	var p Post	// claim a var, empty now, default value

	if err := decoder.Decode(&p); err != nil {	// acquire data (r.Body, which is the data user want to post),
							// then cast it into JSON format
							// base on the Post struct parsing, setter code here is much less
							// compare with Java code here, see doc
		panic(err)	// handle error
		return
	}

	fmt.Fprintf(w, "Post received: %s\n", p.Message)
}

/*
// a simple handlerSearch example
func handlerSearch(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one request for search")
	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")

	fmt.Fprintf(w, "Search received: %s %s", lat, lon)
}
*/

func handlerSearch(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one request for search")
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)    // cast String into 64bit float
	lon, _ := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	// range is optional
	ran := DISTANCE     // define range as const
	if val := r.URL.Query().Get("range"); val != "" {     // if distance is passed para, then get
		ran = val + "km"
	}

	fmt.Printf("Search received: %f %f %s", lat, lon, ran)

	// Return a fake post
	// later on, data should be retrieved from database
	p := &Post{	// init a post, p get its reference, like C++
			// claim a reference p as fake post, which in fact should be retrieve from DB
			// keep this reference, later on cast it into JSON
		// init, fake a one on one mapping (later, cast it into JSON)
		User:"1111",
		Message:"一生必去的100个地方",
		Location: Location{
			Lat:lat,
			Lon:lon,
		},
	}

	js, err := json.Marshal(p)    // xuliehua, cast into JSON format !!!
	if err != nil {
		panic(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}


