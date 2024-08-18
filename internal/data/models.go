package data

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

const (
	dbFilePath = ".maggi"
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
	var db *sql.DB
	var err error
	db, err = sql.Open("sqlite3", dbFileName)
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
