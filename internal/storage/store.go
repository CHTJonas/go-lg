package storage

import (
	"strings"

	"github.com/CHTJonas/go-lg/internal/logging"
	"github.com/dgraph-io/badger"
)

type Store struct {
	db *badger.DB
}

const loggingPrefix string = "db"

func NewStore(path string, level logging.Level) *Store {
	store := &Store{}
	logger := logging.NewPrefixedLogger(loggingPrefix, level)
	opts := badger.DefaultOptions(path).WithLogger(logger)
	logger.Infof("Opening database at %s...", path)
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

func (store *Store) TrimWrite(prefix string, data []byte) (string, error) {
	str := string(data)
	trm := strings.TrimSpace(str)
	return store.Write(prefix, []byte(trm))
}
