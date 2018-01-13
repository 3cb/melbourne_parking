package main

import (
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

func updateDB(db *bolt.DB, msg []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Spaces"))
		err := b.Put([]byte("current"), msg)
		if err != nil {
			return err
		}
		return nil
	})
}

func queryDB(db *bolt.DB) []byte {
	msg := []byte{}
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Spaces"))
		v := b.Get([]byte("current"))
		msg = append(msg, v...)
		return nil
	})
	return msg
}
