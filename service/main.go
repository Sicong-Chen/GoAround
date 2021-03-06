package main

import (
	elastic "gopkg.in/olivere/elastic.v3"
	"fmt"
	"net/http"
	"encoding/json"
	"log"
	"strconv"
	"reflect"
	"context"
	"cloud.google.com/go/bigtable"
	"github.com/pborman/uuid"
)

// 9/16/2017
// IP is a AWS EC2 Image IP
const (
	INDEX = "around"
	TYPE = "post"
	DISTANCE = "200km"

	// start a GCE instance -> to use ElasticSearch -> get PROJECT_ID
	// In this case, ES is implemented by AWS EC2
	PROJECT_ID = "analog-ship-179904"
	// set a BigTable instance in GCP
	BT_INSTANCE = "around-post"
	// deploy it to AWS EC2, use ElasticSearch(which is on GAE - one of the Go service) on EC2
	ES_URL = "http://34.213.92.223:9200"

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
	// Create a client
	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
		return
	}

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(INDEX).Do()
	if err != nil {
		panic(err)
	}
	if !exists {
		// Create a new index.
		mapping := `{
                    "mappings":{
                           "post":{
                                  "properties":{
                                         "location":{
                                                "type":"geo_point"
                                         }
                                  }
                           }
                    }
             }
             `
		_, err := client.CreateIndex(INDEX).Body(mapping).Do()
		if err != nil {
			// Handle error
			panic(err)
		}
	}

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

	/*
	// Example: Add a document to the index
	tweet := Tweet{User: "olivere", Message: "Take Five"}
	_, err = client.Index().
		Index("twitter").
		Type("tweet").
		Id("1").
		BodyJson(tweet).
		Refresh(true).
		Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	*/


	// Create a client
	es_client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
		return
	}

	id := uuid.New()	//  generate  random number

	// Save it to index
	_, err = es_client.Index().
		Index(INDEX).
		Type(TYPE).
		Id(id).
		BodyJson(p).
		Refresh(true).
		Do()
	if err != nil {
		panic(err)
		return
	}

	fmt.Printf("Post is saved to Index: %s\n", p.Message)
	// fmt.Fprintf(w, "Post to Index received: %s\n", p.Message)



	// Post to BigTable
	ctx := context.Background()
	// update project name here
	bt_client, err := bigtable.NewClient(ctx, PROJECT_ID, BT_INSTANCE)
	if err != nil {
		panic(err)
		return
	}

	/*
	Based on what Google document, how to save our Post into BT?
	tbl := client.Open("mytable")
	mut := bigtable.NewMutation()
	mut.Set("links", "maps.google.com", bigtable.Now(), []byte("1"))
	mut.Set("links", "golang.org", bigtable.Now(), []byte("1"))
	err := tbl.Apply(ctx, "com.google.cloud", mut)
	*/



	// TODO: save Post into BT as well
	tbl := bt_client.Open("post")
	mut := bigtable.NewMutation()
	t := bigtable.Now()

	// column family - post
	mut.Set("post", "user", t, []byte(p.User))
	mut.Set("post", "message", t, []byte(p.Message))
	// column family - location
	mut.Set("location", "lat", t, []byte(strconv.FormatFloat(p.Location.Lat, 'f', -1, 64)))
	mut.Set("location", "lon", t, []byte(strconv.FormatFloat(p.Location.Lon, 'f', -1, 64)))

	err = tbl.Apply(ctx, id, mut)
	if err != nil {
		panic(err)
		return
	}
	fmt.Printf("Post is saved to BigTable: %s\n", p.Message)
	// fmt.Fprintf(w, "Post to BigTable received: %s\n", p.Message)


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

/*
 * handlerSearch for 19th
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
*/


// use KD tree to check search result, see if it's within 200km of the post position
func handlerSearch(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one request for search")
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lon, _ := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)

	// range is optional
	ran := DISTANCE
	if val := r.URL.Query().Get("range"); val != "" {
		ran = val + "km"
	}

	// fmt.Printf(w, "Search received: %f %f %s", lat, lon, ran)
	// error: cannot use w (type http.ResponseWriter) as type string in argument to fmt.Printf

	fmt.Printf("Search received: %f %f %s", lat, lon, ran)

	// below is different
	// use GeoIndex to implement KD tree ALG to do diatance range filter

	// Create a client
	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
		return
	}

	// Define geo distance query as specified in
	// https://www.elastic.co/guide/en/elasticsearch/reference/5.2/query-dsl-geo-distance-query.html
	q := elastic.NewGeoDistanceQuery("location")
	q = q.Distance(ran).Lat(lat).Lon(lon)   // from IOS front end

	// Some delay may range from seconds to minutes.
	searchResult, err := client.Search().
		Index(INDEX).
		Query(q).
		Pretty(true).
		Do()
	if err != nil {
		// Handle error
		panic(err)
	}

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch.
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
	// TotalHits is another convenience function that works even when something goes wrong.
	fmt.Printf("Found a total of %d post\n", searchResult.TotalHits())

	// Each is a convenience function that iterates over hits in a search result.
	// It makes sure you don't need to check for nil values in the response.
	// However, it ignores errors in serialization.
	var typ Post
	var ps []Post
	for _, item := range searchResult.Each(reflect.TypeOf(typ)) {
		p := item.(Post)
		fmt.Printf("Post by %s: %s at lat %v and lon %v\n", p.User, p.Message, p.Location.Lat, p.Location.Lon)
		// TODO(vincent): Perform filtering based on keywords such as web spam etc.
		ps = append(ps, p)

	}
	js, err := json.Marshal(ps)	// serialize into JSON
	if err != nil {
		panic(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}


