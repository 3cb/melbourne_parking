package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/3cb/ssc"
	"github.com/boltdb/bolt"
)

func poll(db *bolt.DB, pool *ssc.SocketPool) {
	ticker := time.NewTicker(time.Minute * 1)
	loc, _ := time.LoadLocation("Australia/Melbourne")

	spots, err := updateSpots()
	if err != nil {
		log.Printf("Unable to update spots data: %v", err)
	}
	t := strings.Split(fmt.Sprint(time.Now().In(loc)), " ")
	buf := serialize(spots, fmt.Sprintf("%v %v", t[0], t[1]))
	err = updateDB(db, buf)
	if err != nil {
		log.Printf("Unable to store spots data in database: %v", err)
	}
	pool.Pipes.InboundBytes <- ssc.Data{Type: 2, Payload: buf}

	for {
		<-ticker.C

		spots, err := updateSpots()
		if err != nil {
			log.Printf("Unable to update spots data: %v", err)
		}
		t := strings.Split(fmt.Sprint(time.Now().In(loc)), " ")
		buf := serialize(spots, fmt.Sprintf("%v %v", t[0], t[1]))
		err = updateDB(db, buf)
		if err != nil {
			log.Printf("Unable to store spots data in database: %v", err)
		}
		pool.Pipes.InboundBytes <- ssc.Data{Type: 2, Payload: buf}

	}
}

func updateSpots() ([]Spot, error) {
	spots := []Spot{}

	resp, err := http.Get("https://data.melbourne.vic.gov.au/resource/vh2v-4nfs.json")
	if err != nil {
		return spots, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return spots, err
	}
	err = json.Unmarshal(data, &spots)
	if err != nil {
		return spots, err
	}
	return spots, nil
}
