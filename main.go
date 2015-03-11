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
	rename   bool
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
	flag.BoolVar(&rename, "rename", false, "place as a rename instead of a insertion")
	flag.Parse()
}

func normalizeArgs() {
	if (srcKey != "" && srcJSON != "") || (srcKey == "" && srcJSON == "") {
		log.Fatal("Either the source key or JSON flag must utilized")
	} else if (destKey != "" && destJSON != "") || (destKey == "" && destJSON == "") {
		log.Fatal("Either the destination key or JSON flag must utilized")
	}
}

// readJSONFile constructs a tree from a specifed JSON file. The function exits if the
// file is not found.
func readJSONFile(filename string) tree {
	values := tree{}

	// open and read file data
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Failed to open srcJSON file => {%s}", err)
	}

	// write data into tree
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&values)
	if err != nil {
		log.Printf("Failed to decode json in file => {%s}", err)
	}

	return values
}

// writeJSONFile writes retrieved data to a file.
func writeJSONFile(t tree, filename string) {
	// marshal data retrieved into JSON
	data, err := json.Marshal(t)
	if err != nil {
		log.Fatalf("Error marshaling data for JSON => {%s}", err)
	}

	// write data into file
	err = ioutil.WriteFile(filename, data, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to write json data to file, %s => {%s}", destJSON, err)
	}
}

// readConsulTree constructs a tree from the Consul KV store at the specified key.
func readConsulTree(key string) tree {
	values := tree{}

	// try to find values in key given, else take all values
	pairs, _, err := kv.List(srcKey, &consul.QueryOptions{})
	if err != nil {
		log.Fatalf("Error retrieving data for specified key, %s => {%s}", srcKey, err)
	} else if len(pairs) == 0 {
		log.Fatalf("Failed to find any data, %s", srcKey)
	}

	values.build(pairs)

	return values
}

// putConsulTree adds a config tree to a consul KV store at the specified key.
func putConsulTree(t tree, key string) {
	if !rename {
		t.update("/" + key)
		return
	}

	for k, v := range t {
		subTree, ok := v.(map[string]interface{})
		if ok {
			// push retrieved data to a Consul key
			tree(subTree).update("/" + key)
		} else {
			push(key+"/"+k, v)
		}
	}
}

func main() {
	values := tree{}
	normalizeArgs()

	// 1. find the input data from either a file or Consul key
	if srcJSON != "" {
		values = readJSONFile(srcJSON)
	} else {
		values = readConsulTree(srcKey)
	}

	// 2. write the src data to the destination
	if destJSON != "" {
		writeJSONFile(values, destJSON)
	} else {
		putConsulTree(values, destKey)
	}
}
