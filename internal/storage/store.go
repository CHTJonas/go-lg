package storage

import (
	"github.com/dgraph-io/badger"
)

type Store struct {
	db *badger.DB
}

func NewStore(path string) *Store {
	s := &Store{}
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		panic(err)
	}
	s.db = db
	return s
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) Read(prefix string, uid string) []byte {
	var data []byte
	s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(prefix + uid))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			data = append([]byte{}, val...)
			return nil
		})
	})
	return data
}

func (s *Store) Write(prefix string, data []byte) (string, error) {
	uid := GenerateUID()
	return uid, s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(prefix+uid), data)
	})
}
