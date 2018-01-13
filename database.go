package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/boltdb/bolt"
)

func createBucket(db *bolt.DB, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return err
		}
		return nil
	})
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
