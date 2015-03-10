package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"

	consul "github.com/hashicorp/consul/api"
)

var (
	kv       *consul.KV
	srcKey   string
	srcJSON  string
	destKey  string
	destJSON string
)

// init handles connecting to the Consul instance and reading in the flags.
func init() {
	// NOTE: this will utilize CONSUL_HTTP_ADDR if it is set.
	client, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to connect to Consul => {%s}", err)
	}
	kv = client.KV()

	flag.StringVar(&srcKey, "srcKey", "", "key to move values from")
	flag.StringVar(&destKey, "destKey", "", "key to move values to")
	flag.StringVar(&srcJSON, "srcJSON", "", "file to import values from")
	flag.StringVar(&destJSON, "destJSON", "", "file to export values to")
	flag.Parse()

	if (srcKey != "" && srcJSON != "") || (srcKey == "" && srcJSON == "") {
		log.Fatal("Either the source key or JSON flag must utilized")
	} else if (destKey != "" && destJSON != "") || (destKey == "" && destJSON == "") {
		log.Fatal("Either the destination key or JSON flag must utilized")
	}
}

func main() {
	values := tree{}

	// 1. find the input data from either a file or Consul key
	if srcJSON != "" {
		// open and read file data
		file, err := os.Open(srcJSON)
		if err != nil {
			log.Printf("Failed to open srcJSON file => {%s}", err)
		}

		// write data into tree
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&values)
		if err != nil {
			log.Printf("Failed to decode json in file => {%s}", err)
		}

		log.Printf("%#v", values)
	} else {
		// try to find values in key given, else take all values
		pairs, _, err := kv.List(srcKey, &consul.QueryOptions{})
		if err != nil {
			log.Fatalf("Error retrieving data for specified key, %s => {%s}", srcKey, err)
		} else if len(pairs) == 0 {
			log.Fatalf("Failed to find any data, %s", srcKey)
		}

		values.build(pairs)
	}

	// 2. write the src data to the destination
	if destJSON != "" {
		// write retrieved data to a file

		// marshal data retrieved into JSON
		data, err := json.Marshal(values)
		if err != nil {
			log.Fatalf("Error marshaling data for JSON => {%s}", err)
		}

		err = ioutil.WriteFile(destJSON, data, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to write json data to file, %s => {%s}", destJSON, err)
		}
	} else {
		for _, v := range values {
			subTree, ok := v.(map[string]interface{})
			if ok {
				// push retrieved data to a Consul key
				tree(subTree).update("/" + destKey)
			} else {
				log.Fatal("Consul Loader does not support root level keys")
			}
		}
	}
}
