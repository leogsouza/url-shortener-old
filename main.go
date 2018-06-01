package main

import (
	"log"
	"net/http"
	"os"

	"github.com/couchbase/gocb"
	"github.com/gorilla/mux"
)

var bucket *gocb.Bucket
var bucketName string

func ExpandEndpoint(w http.ResponseWriter, r *http.Request) {}

func CreateEndpoint(w http.ResponseWriter, r *http.Request) {}

func RootEndpoint(w http.ResponseWriter, r *http.Request) {}

func main() {
	router := mux.NewRouter()
	cluster, err := gocb.Connect("couchbase://" + os.Getenv("DATABASE_HOST"))
	bucketName := os.Getenv("DATABASE_NAME")
	bucket, err := cluster.OpenBucket(bucketName, os.Getenv("DATABASE_PASSWORD"))

	router.HandleFunc("/{id}", RootEndpoint).Methods("GET")
	router.HandleFunc("/expand", ExpandEndpoint).Methods("GET")
	router.HandleFunc("/create", CreateEndpoint).Methods("PUT")
	log.Fatal(http.ListenAndServe(":3333", router))
}
