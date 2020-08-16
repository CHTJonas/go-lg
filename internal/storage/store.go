package storage

import (
	"github.com/chtjonas/go-lg/internal/logging"
	"github.com/dgraph-io/badger"
)

type Store struct {
	db *badger.DB
}

func NewStore(path string) *Store {
	store := &Store{}
	logger := logging.NewPrefixedLogger("db")
	opts := badger.DefaultOptions(path).WithLogger(logger)
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	store.db = db
	return store
}

func (store *Store) Close() {
	store.db.Close()
}

func (store *Store) Read(prefix string, uid string) []byte {
	var data []byte
	store.db.View(func(txn *badger.Txn) error {
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

func (store *Store) Write(prefix string, data []byte) (string, error) {
	uid := GenerateUID()
	return uid, store.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(prefix+uid), data)
	})
}
