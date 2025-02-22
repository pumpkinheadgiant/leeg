package migration

import (
	"log/slog"
	"strconv"

	"go.etcd.io/bbolt"
	"phg.com/leeg/svc"
)

type Migrator struct{}

type migrationFunc func(tx *bbolt.Tx) error

func (m Migrator) Migrate(db *bbolt.DB) error {
	slog.Info("performing migrations")
	return db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(metaBucketKey))
		if err != nil {
			return err
		}
		versionValue := bucket.Get([]byte(dbVersionKey))
		version, _ := strconv.Atoi(string(versionValue))
		migrations := m.getMigrations()
		for ; version < len(migrations); version++ {

			if err := migrations[version](tx); err != nil {
				return err
			}

			newVersionValue := strconv.Itoa(version + 1)
			if err := bucket.Put([]byte(dbVersionKey), []byte(newVersionValue)); err != nil {
				return err
			}
		}
		slog.Info("migrations complete", "version", version)

		return nil
	})
}

func (m Migrator) getMigrations() []migrationFunc {
	return []migrationFunc{
		// Migration 1
		func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(svc.LeegsBucketKey))
			if err != nil {
				return err
			}
			return nil
		},
	}
}

const metaBucketKey = "meta"
const dbVersionKey = "version"
