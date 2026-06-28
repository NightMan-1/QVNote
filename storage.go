package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

const boltDBFileName = "qvnote.db"

var (
	configBucket    = []byte("config")
	notebookBucket  = []byte("notebook")
	noteBucket      = []byte("note")
	tagsBucket      = []byte("tags")
	favoritesBucket = []byte("favorites")
)

// Store управляет единым файлом bbolt и предоставляет доступ к логическим KV-базам.
type Store struct {
	db *bolt.DB
}

// KVStore — обёртка над одним bucket bbolt.
type KVStore struct {
	bucket []byte
	db     *bolt.DB
}

// OpenStore открывает (или создаёт) файл bbolt в dataDir и инициализирует bucket'ы.
func OpenStore(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0750); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	path := filepath.Join(dataDir, boltDBFileName)
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bbolt: %w", err)
	}

	s := &Store{db: db}
	if err := s.createBuckets(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close закрывает файл базы данных.
func (s *Store) Close() error {
	return s.db.Close()
}

// DB возвращает низкоуровневый handle bbolt (нужен для batch-операций миграции).
func (s *Store) DB() *bolt.DB {
	return s.db
}

// Config возвращает хранилище конфигурации.
func (s *Store) Config() *KVStore { return s.Bucket(configBucket) }

// NoteBook возвращает хранилище блокнотов.
func (s *Store) NoteBook() *KVStore { return s.Bucket(notebookBucket) }

// Note возвращает хранилище заметок.
func (s *Store) Note() *KVStore { return s.Bucket(noteBucket) }

// Tags возвращает хранилище тегов.
func (s *Store) Tags() *KVStore { return s.Bucket(tagsBucket) }

// Favorites возвращает хранилище избранного.
func (s *Store) Favorites() *KVStore { return s.Bucket(favoritesBucket) }

// Bucket возвращает KVStore для произвольного bucket'а.
func (s *Store) Bucket(name []byte) *KVStore {
	return &KVStore{bucket: name, db: s.db}
}

func (s *Store) createBuckets() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, name := range [][]byte{configBucket, notebookBucket, noteBucket, tagsBucket, favoritesBucket} {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return fmt.Errorf("create bucket %q: %w", name, err)
			}
		}
		return nil
	})
}

// Get возвращает значение по ключу. Копия безопасна за пределами транзакции.
func (kv *KVStore) Get(key []byte) ([]byte, error) {
	var value []byte
	err := kv.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(kv.bucket)
		if b == nil {
			return errors.New("bucket not found")
		}
		if v := b.Get(key); v != nil {
			value = append([]byte{}, v...)
		}
		return nil
	})
	return value, err
}

// Set сохраняет ключ/значение.
func (kv *KVStore) Set(key, value []byte) error {
	return kv.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(kv.bucket).Put(key, value)
	})
}

// Del удаляет ключ.
func (kv *KVStore) Del(key []byte) error {
	return kv.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(kv.bucket).Delete(key)
	})
}

// Exists проверяет наличие ключа.
func (kv *KVStore) Exists(key []byte) (bool, error) {
	var exists bool
	err := kv.db.View(func(tx *bolt.Tx) error {
		exists = tx.Bucket(kv.bucket).Get(key) != nil
		return nil
	})
	return exists, err
}

// Scan проходит по всем записям bucket'а. Передаются копии ключей и значений.
func (kv *KVStore) Scan(fn func(key, value []byte) error) error {
	return kv.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(kv.bucket).ForEach(func(k, v []byte) error {
			keyCopy := append([]byte{}, k...)
			valCopy := append([]byte{}, v...)
			return fn(keyCopy, valCopy)
		})
	})
}

// Keys перебирает только ключи.
func (kv *KVStore) Keys(fn func(key []byte) error) error {
	return kv.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(kv.bucket).ForEach(func(k, _ []byte) error {
			return fn(append([]byte{}, k...))
		})
	})
}

// Flush удаляет все записи в bucket'е (пересоздаёт bucket).
func (kv *KVStore) Flush() error {
	return kv.db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket(kv.bucket); err != nil {
			return fmt.Errorf("delete bucket: %w", err)
		}
		if _, err := tx.CreateBucket(kv.bucket); err != nil {
			return fmt.Errorf("recreate bucket: %w", err)
		}
		return nil
	})
}
