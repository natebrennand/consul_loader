package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

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

// tree is a structure used to build a representation of the consul config.
type tree map[string]interface{}

// String returns a string representation of the Tree.
func (t tree) String() (repr string) {
	for k, v := range t {
		subTree, ok := v.(tree)
		if ok {
			repr += fmt.Sprintf("%s: {%s},\n", k, subTree.String())
		} else {
			repr += fmt.Sprintf("%s: %s\n", k, v)
		}
	}
	return
}

// add traverses the tree from the split key to find the proper place to put the value.
func (t tree) add(k string, v interface{}) {
	if k == "" {
		return
	}
	path := strings.Split(k, "/")

	if len(path) == 1 {
		t[k] = v
	} else {
		subKey := path[0]
		subTree, exists := t[subKey]
		if !exists {
			t[subKey] = tree{}
			subTree = t[subKey]
		}

		subTree.(tree).add(strings.Join(path[1:], "/"), v)
	}
}

// build adds a series of KVPairs to the tree.
func (t tree) build(kvs consul.KVPairs) {
	for _, pair := range kvs {
		// jank
		if destJSON != "" {
			t.add(pair.Key, string(pair.Value))
		} else {
			t.add(pair.Key, pair.Value)
		}
	}
}

func (t tree) update(base string) {
	for k, v := range t {
		subTree, ok := v.(tree)
		if ok {
			subTree.update(base + "/" + k)
		} else {
			log.Printf("%s => %s", base+"/"+k, string(v.([]byte)))
			_, err := kv.Put(&consul.KVPair{
				Key:   (base + "/" + k)[1:],
				Value: v.([]byte),
			}, nil)
			if err != nil {
				log.Fatalf("Failed to write to Consul => {%s}", err)
			}
		}
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
			subTree, ok := v.(tree)
			if ok {
				// push retrieved data to a Consul key
				subTree.update("/" + destKey)
			} else {
				log.Fatal("Consul Loader does not support root level keys")
			}
		}
	}
}
