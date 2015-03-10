package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	consul "github.com/hashicorp/consul/api"
)

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

func resolveBytes(v interface{}) []byte {
	switch val := v.(type) {
	case []byte:
		return val
	case string:
		return []byte(val)
	case int:
		return []byte(strconv.Itoa(val))
	default:
		log.Fatal("Unsupported type, please file an issue")
	}

	return []byte{}
}

func (t tree) update(base string) {
	for k, v := range t {
		subTree, ok := v.(map[string]interface{})
		if ok {
			tree(subTree).update(base + "/" + k)
		} else {
			val := resolveBytes(v)

			log.Printf("%s => %s", base+"/"+k, val)
			_, err := kv.Put(&consul.KVPair{
				Key:   (base + "/" + k)[1:],
				Value: val,
			}, nil)
			if err != nil {
				log.Fatalf("Failed to write to Consul => {%s}", err)
			}
		}
	}
}
