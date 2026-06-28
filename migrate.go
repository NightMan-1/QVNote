package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	lediscfg "github.com/ledisdb/ledisdb/config"
	"github.com/ledisdb/ledisdb/ledis"
	bolt "go.etcd.io/bbolt"
)

// migrateFromLedisIfNeeded выполняет одноразовый перенос данных из старой LedisDB в bbolt.
// Старая папка goleveldb_data переименовывается, чтобы миграция не запускалась повторно.
func migrateFromLedisIfNeeded(dataDir string, store *Store) error {
	legacyDir := filepath.Join(dataDir, "goleveldb_data")
	if info, err := os.Stat(legacyDir); os.IsNotExist(err) || !info.IsDir() {
		fmt.Printf("No legacy LedisDB data found at %s, skipping migration.\n", legacyDir)
		return nil
	}

	boltPath := filepath.Join(dataDir, boltDBFileName)
	if _, err := os.Stat(boltPath); err == nil {
		// Файл bbolt уже существует. Мигрируем только если он пустой,
		// иначе считаем, что миграция уже выполнена или пользователь начал с чистой базы.
		empty, err := isBoltDBEmpty(store)
		if err != nil {
			return fmt.Errorf("check bbolt emptiness: %w", err)
		}
		if !empty {
			fmt.Printf("bbolt database %s already exists and is not empty, skipping migration.\n", boltPath)
			return nil
		}
		fmt.Printf("bbolt database %s exists but is empty, will migrate legacy data.\n", boltPath)
	}

	fmt.Printf("Legacy LedisDB data found at %s, migrating to bbolt...\n", legacyDir)

	cfg := lediscfg.NewConfigDefault()
	cfg.DataDir = dataDir
	oldDB, err := ledis.Open(cfg)
	if err != nil {
		return fmt.Errorf("open legacy ledisdb: %w", err)
	}
	defer oldDB.Close()
	fmt.Println("Legacy LedisDB opened successfully.")

	mappings := []struct {
		idx    int
		bucket []byte
		name   string
	}{
		{0, configBucket, "config"},
		{1, notebookBucket, "notebook"},
		{2, noteBucket, "note"},
		{3, tagsBucket, "tags"},
		{4, favoritesBucket, "favorites"},
	}

	for _, m := range mappings {
		oldKV, err := oldDB.Select(m.idx)
		if err != nil {
			return fmt.Errorf("select ledis db %d (%s): %w", m.idx, m.name, err)
		}
		count, err := migrateBucket(oldKV, store.DB(), m.bucket)
		if err != nil {
			return fmt.Errorf("migrate %s: %w", m.name, err)
		}
		fmt.Printf("  migrated %d records to bucket %s\n", count, m.name)
	}

	migratedName := fmt.Sprintf("goleveldb_data.migrated-%d", time.Now().Unix())
	if err := os.Rename(legacyDir, filepath.Join(dataDir, migratedName)); err != nil {
		return fmt.Errorf("rename legacy data dir: %w", err)
	}

	fmt.Printf("Migration finished. Legacy data kept at %s\n", migratedName)
	return nil
}

func migrateBucket(oldDB *ledis.DB, newDB *bolt.DB, bucketName []byte) (int, error) {
	count := 0
	err := newDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}

		cursor := []byte(nil)
		for {
			keys, err := oldDB.Scan(ledis.KV, cursor, 0, false, "")
			if err != nil || len(keys) == 0 {
				break
			}
			for _, key := range keys {
				cursor = key
				value, err := oldDB.Get(key)
				if err != nil {
					return err
				}
				if err := b.Put(key, value); err != nil {
					return err
				}
				count++
			}
		}
		return nil
	})
	return count, err
}

func isBoltDBEmpty(store *Store) (bool, error) {
	empty := true
	err := store.DB().View(func(tx *bolt.Tx) error {
		for _, name := range [][]byte{configBucket, notebookBucket, noteBucket, tagsBucket, favoritesBucket} {
			b := tx.Bucket(name)
			if b == nil {
				continue
			}
			if b.Stats().KeyN > 0 {
				empty = false
				return nil
			}
		}
		return nil
	})
	return empty, err
}
