package main

import (
	"flag"
	"log"

	"github.com/agneum/gocodelab/api"
)

func main() {
	bindAddr := flag.String("bind_addr", ":8888", "Set bind address")
	flag.Parse()
	a := api.New(*bindAddr)
	log.Fatal(a.Start())
}
