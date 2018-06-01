package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/speps/go-hashids"

	"github.com/couchbase/gocb"
	"github.com/couchbase/gocbcore"
	"github.com/gorilla/mux"
)

// MyURL represents a url shortner structure
type MyURL struct {
	ID       string `json:"id,omitempty"`
	LongURL  string `json:"longUrl,omitempty"`
	ShortUrl string `json:"shortURL,omitempty"`
}

var bucket *gocb.Bucket
var bucketName string

// ExpandEndpoint Finds the url by shortURL and show it
func ExpandEndpoint(w http.ResponseWriter, r *http.Request) {
	var n1qlParams []interface{}
	query := gocb.NewN1qlQuery("SELECT `" + bucketName + "`.* FROM `" +
		bucketName + "` WHERE shortURL = $1")
	params := r.URL.Query()
	n1qlParams = append(n1qlParams, params.Get("shortURL"))
	rows, _ := bucket.ExecuteN1qlQuery(query, n1qlParams)
	var row MyURL
	rows.One(&row)
	json.NewEncoder(w).Encode(row)
}

// CreateEndpoint creates a short url in database if register do not exists
func CreateEndpoint(w http.ResponseWriter, r *http.Request) {
	var url MyURL
	_ = json.NewDecoder(r.Body).Decode(&url)

	var n1qlParams []interface{}
	n1qlParams = append(n1qlParams, url.LongURL)
	log.Println("SELECT `" + bucketName + "`.* FROM `" + bucketName + "` WHERE longUrl = $1")
	query := gocb.NewN1qlQuery("SELECT `" + bucketName + "`.* FROM `" + bucketName + "` WHERE longUrl = $1")
	rows, err := bucket.ExecuteN1qlQuery(query, n1qlParams)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		log.Printf("%v", err)
		return
	}
	var row MyURL
	rows.One(&row)
	if row == (MyURL{}) {
		hd := hashids.NewData()
		h, _ := hashids.NewWithData(hd)
		now := time.Now()
		url.ID, _ = h.Encode([]int{int(now.Unix())})
		url.ShortUrl = "http://localhost:3333/" + url.ID
		bucket.Insert(url.ID, url, 0)
	} else {
		url = row
	}
	json.NewEncoder(w).Encode(url)
}

// RootEndpoint redirects to url found in our database
func RootEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var url MyURL
	bucket.Get(params["id"], &url)
	http.Redirect(w, r, url.LongURL, http.StatusFound)
}

func main() {
	gocbcore.SetLogger(gocbcore.VerboseStdioLogger())
	router := mux.NewRouter()
	var err error
	cluster, err := gocb.Connect("couchbase://" + os.Getenv("DATABASE_HOST"))
	handleError(err)
	cluster.Authenticate(gocb.PasswordAuthenticator{
		Username: os.Getenv("DATABASE_USERNAME"),
		Password: os.Getenv("DATABASE_PASSWORD"),
	})

	bucketName = os.Getenv("DATABASE_NAME")
	bucket, err = cluster.OpenBucket(bucketName, "")
	handleError(err)

	router.HandleFunc("/{id}", RootEndpoint).Methods("GET")
	router.HandleFunc("/expand/", ExpandEndpoint).Methods("GET")
	router.HandleFunc("/create", CreateEndpoint).Methods("PUT")

	log.Fatal(http.ListenAndServe(":3333", router))
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
