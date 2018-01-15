package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/3cb/ssc"
	"github.com/gorilla/mux"

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

	go poll(db, pool)

	// routes
	r := mux.NewRouter()

	r.Handle("/", http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/dist").Handler(http.FileServer(http.Dir("./static")))

	r.Handle("/api/spots", spotsHandler(db))

	upgrader := &websocket.Upgrader{}
	r.Handle("/ws", wsHandler(db, pool, upgrader))

	log.Fatal(http.ListenAndServe(":5050", r))
}

func spotsHandler(db *bolt.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := queryDB(db)
		w.Write(msg)
	})
}

func wsHandler(db *bolt.DB, pool *ssc.SocketPool, upgrader *websocket.Upgrader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := pool.AddClientSocket(upgrader, w, r)
		if err != nil {
			log.Printf("Unable to create new socket connection")
		} else {
			log.Printf("New websocket client connected.")
		}
	})
}
