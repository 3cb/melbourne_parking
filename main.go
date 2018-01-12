package main

import (
	"log"

	"github.com/3cb/ssc"

	"github.com/boltdb/bolt"
)

func main() {
	// start database
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatalf("Unable to open database: %s", err)
	}
	defer db.Close()
	err = createBucket(db, "Spaces")
	if err != nil {
		log.Fatalf("unable to create bucket: %s", err)
	}

	// start websocket pool
	config := ssc.PoolConfig{
		IsReadable: true,
		IsWritable: true,
		IsJSON:     false,
	}
	pool, err := ssc.NewSocketPool(config)
	if err != nil {
		log.Fatalf("Unable to create socket pool: %s", err)
	}

}