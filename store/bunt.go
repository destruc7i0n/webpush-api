package store

import (
	"encoding/json"
	"fmt"

	buntdb "github.com/tidwall/buntdb"
)

type Store struct {
	db *buntdb.DB
}

const (
	IndexSubscriptionTopic = "subscription_topic"
	IndexNotificationTopic = "notification_topic"
)

func NewStore() (*Store, error) {
	// db, err := buntdb.Open(":memory:")
	db, err := buntdb.Open("store.db")
	db.CreateIndex(IndexSubscriptionTopic, fmt.Sprintf("%s:*", KeySubscription), buntdb.IndexJSON(string(KeyTopic)))
	db.CreateIndex(IndexNotificationTopic, fmt.Sprintf("%s:*", KeyNotification), buntdb.IndexJSON(string(KeyTopic)))

	db.Shrink() // compact the database

	if err != nil {
		return nil, err
	}

	return &Store{
		db: db,
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Get(key string) ([]byte, error) {
	var val []byte
	err := s.db.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(key)
		if err != nil {
			return err
		}
		val = []byte(v)
		return nil
	})
	return val, err
}

func (s *Store) Set(key string, value []byte) error {
	return s.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key, string(value), nil)
		return err
	})
}

func (s *Store) Delete(key string) error {
	// log.Printf("deleting key %s", key)
	return s.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(key)
		return err
	})
}

func (s *Store) SetStruct(key string, value interface{}) error {
	// encode the value
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// set the value
	return s.Set(key, val)
}

func (s *Store) GetStruct(key string, value interface{}) error {
	// get the value
	val, err := s.Get(key)
	if err != nil {
		return err
	}

	// decode the value
	return json.Unmarshal(val, value)
}

func (s *Store) AscendBy(prefix string) (map[string]string, error) {
	list := make(map[string]string)
	err := s.db.View(func(tx *buntdb.Tx) error {
		err := tx.AscendKeys(prefix, func(key, value string) bool {
			list[key] = value
			return true
		})
		return err
	})
	return list, err
}
