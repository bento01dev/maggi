package data

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbFilePath = ".maggi/"
	dbFileName = "maggi.db"
)

const (
	profileDDL = `
    CREATE TABLE IF NOT EXISTS profiles (
    id INTEGER NOT NULL PRIMARY KEY,
    name STRING NOT NULL
    );`
)

type DataModel struct {
	Profiles *Profiles
}

func NewDataModel(db *sql.DB) DataModel {
	return DataModel{
		Profiles: &Profiles{db: db},
	}
}

func Setup() (*sql.DB, error) {
	var err error
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	maggiLoc := filepath.Join(homeDir, dbFilePath)
    _, err = os.Stat(maggiLoc)
    if os.IsNotExist(err) {
        err = os.MkdirAll(maggiLoc, 0755)
        if err != nil {
            return nil, err
        }
    }

    dbPath := filepath.Join(maggiLoc, dbFileName)

	var db *sql.DB
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return db, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(profileDDL); err != nil {
		return nil, err
	}

	return db, nil
}
