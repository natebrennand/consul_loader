package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	consul "github.com/hashicorp/consul/api"
)

var (
	kv       *consul.KV
	srcKey   string
	destKey  string
	srcJSON  string
	destJSON string
)

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
}

type tree map[string]interface{}

func (t tree) String() string {
	var repr string

	for k, v := range t {
		subTree, ok := v.(tree)
		if ok {
			repr += fmt.Sprintf("%s: {%s},\n", k, subTree.String())
		} else {
			repr += fmt.Sprintf("%s: %s\n", k, v)
		}
	}
	return repr
}

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

func (t tree) build(kvs consul.KVPairs) {
	for _, pair := range kvs {
		t.add(pair.Key, pair.Value)
	}
}

func main() {
	values := tree{}

	// try to find values in file
	if srcJSON != "" {
		file, err := os.Open(srcJSON)
		if err != nil {
			log.Printf("Failed to open srcJSON file => {%s}", err)
		}

		decoder := json.NewDecoder(file)
		err = decoder.Decode(&values)
		if err != nil {
			log.Printf("Failed to decode json in file => {%s}", err)
		}
	} else { // try to find values in key given, else take all values
		pairs, _, err := kv.List(srcKey, &consul.QueryOptions{})
		if err != nil {
			log.Fatalf("Error retrieving data for specified key, %s => {%s}", srcKey, err)
		} else if len(pairs) == 0 {
			log.Fatalf("Failed to find any data, %s", srcKey)
		}

		values.build(pairs)
	}

	data, err := json.Marshal(values)
	if err != nil {
		log.Fatalf("Error marshaling data for JSON => {%s}", err)
	}

	fmt.Println(string(data))

	// TODO: add dest stuff
	// TODO: decode the values
}
