package dedupe

import (
	"github.com/apex/log"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"os"
)

type Interface interface {
	SeenBefore(id string) bool
	Close()
}

type LevelDb struct {
	db   *leveldb.DB
	path string
}

func (d *LevelDb) SeenBefore(id string) bool {
	byteId := []byte(id)

	if has, err := d.db.Has(byteId, nil); err != nil {
		log.WithError(err).Fatal("LevelDB.Has() method failed")
	} else if has {
		return true
	}

	if err := d.db.Put(byteId, nil, nil); err != nil {
		log.WithError(err).Fatal("LevelDB.Put() method failed")
	}

	return false
}

func (d *LevelDb) Close() {
	d.db.Close()
	if err := os.RemoveAll(d.path); err != nil {
		logger := log.WithError(err).WithField("path", d.path)
		logger.Error("Failed to remove LevelDB temp directory")
	}
}

func New() *LevelDb {
	tmpDir, err := ioutil.TempDir("", "stripe-leveldb")
	if err != nil {
		log.WithError(err).Fatal("could not create temporary dir for leveldb")
	}

	db, err := leveldb.OpenFile(tmpDir, nil)
	if err != nil {
		log.WithError(err).Fatal("could not create leveldb")
	}

	return &LevelDb{db: db, path: tmpDir}
}
