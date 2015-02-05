package main

import (
	"log"

	consul "github.com/hashicorp/consul/api"
)

var (
	client *consul.Client
)

func init() {
	var err error
	client, err = consul.NewClient(consul.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to connect to Consul => {%s}", err)
	}
}

func main() {
}
