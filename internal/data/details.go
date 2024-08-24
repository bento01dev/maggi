package data

import "database/sql"

type DetailType string

const (
	AliasDetail DetailType = "alias"
	EnvDetail   DetailType = "env"
)

type Detail struct {
	ID         int
	Key        string
	Value      string
	DetailType DetailType
	ProfileID  int
}

type Details struct {
	db *sql.DB
}

func (d *Details) GetAll(profileId int) ([]Detail, error) {
	return []Detail{}, nil
}

func (d *Details) Add(key string, value string, detailType DetailType, profileID int) (Detail, error) {
	return Detail{}, nil
}

func (d *Details) Update(detail Detail, key string, value string) (Detail, error) {
	return Detail{}, nil
}

func (d *Details) Delete(detail Detail) error {
	return nil
}

func (d *Details) DeleteAll(profileID int) error {
	return nil
}
