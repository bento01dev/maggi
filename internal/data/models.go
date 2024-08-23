package data

import (
	"context"
	"database/sql"
	"fmt"
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
    BEGIN;
    CREATE TABLE IF NOT EXISTS profiles (
    id INTEGER NOT NULL PRIMARY KEY,
    name STRING NOT NULL
    );
    CREATE INDEX IF NOT EXISTS profile_name_idx ON profiles (name);
    COMMIT;`
	detailsDDL = `
    BEGIN;
    CREATE TABLE IF NOT EXISTS details (
    id INTEGER NOT NULL PRIMARY KEY,
    key STRING NOT NULL,
    value STRING NOT NULL,
    type STRING CHECK( type IN ('alias', 'env') ) NOT NULL,
    profile_id INTEGER NOT NULL,
    FOREIGN KEY(profile_id) REFERENCES profiles(id)
    );
    CREATE INDEX IF NOT EXISTS details_profile_idx ON details (profile_id);
    CREATE INDEX IF NOT EXISTS details_type_idx ON details (type);
    COMMIT;`
)

type DataModel struct {
	Profiles *Profiles
	Details  *Details
}

func NewDataModel(db *sql.DB) DataModel {
	return DataModel{
		Profiles: &Profiles{db: db},
		Details:  &Details{db: db},
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
	pathWithParams := fmt.Sprintf("file:%s?_foreign_keys=true", dbPath)

	var db *sql.DB
	db, err = sql.Open("sqlite3", pathWithParams)
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

	if _, err := db.Exec(detailsDDL); err != nil {
		return nil, err
	}

	return db, nil
}
