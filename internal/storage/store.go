package storage

import (
	"log"
	"strings"

	"github.com/dgraph-io/badger"
)

type Store struct {
	db *badger.DB
}

func NewStore(path string) (*Store, error) {
	store := &Store{}
	opts := badger.DefaultOptions(path).WithLogger(&DBLogger{INFO})
	log.Printf("Opening database at %s...\n", path)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	store.db = db
	return store, nil
}

func (store *Store) Close() error {
	return store.db.Close()
}

func (store *Store) Read(prefix string, uid string) (string, error) {
	var data []byte
	err := store.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(prefix + uid))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			data = append([]byte{}, val...)
			return nil
		})
	})
	return string(data), err
}

func (store *Store) Write(prefix string, data []byte) (string, error) {
	var uid string
	err := store.db.Update(func(txn *badger.Txn) error {
		for {
			uid = GenerateUID()
			_, err := txn.Get([]byte(prefix + uid))
			if err == badger.ErrKeyNotFound {
				break
			} else if err != nil {
				return err
			}
		}
		return txn.Set([]byte(prefix+uid), data)
	})
	return uid, err
}

func (store *Store) TrimWrite(prefix string, data []byte) (string, error) {
	str := string(data)
	trm := strings.TrimSpace(str)
	return store.Write(prefix, []byte(trm))
}
