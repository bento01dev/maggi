package data

import "database/sql"

type DetailType string

func (d DetailType) String() string {
	return string(d)
}

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
	stmt := "SELECT id, key, value, type, profile_id FROM details WHERE profile_id = ?;"
	rows, err := d.db.Query(stmt, profileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	details := []Detail{}
	for rows.Next() {
		detail := &Detail{}
		var typeStr string
		err = rows.Scan(&detail.ID, &detail.Key, &detail.Value, &typeStr, &detail.ProfileID)
		if err != nil {
			return nil, err
		}
		switch typeStr {
		case "alias":
			detail.DetailType = AliasDetail
		case "env":
			detail.DetailType = EnvDetail
		}
		details = append(details, *detail)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return details, nil
}

func (d *Details) Add(key string, value string, detailType DetailType, profileID int) (*Detail, error) {
	var detail Detail
	stmt := "INSERT INTO details (key, value, type, profile_id) VALUES (?, ?, ?, ?);"
	res, err := d.db.Exec(stmt, key, value, detailType.String(), profileID)
	if err != nil {
		return nil, err
	}
	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return nil, err
	}
	detail = Detail{ID: int(id), Key: key, Value: value, DetailType: detailType, ProfileID: profileID}
	return &detail, nil
}

func (d *Details) Update(detail Detail, key string, value string) (*Detail, error) {
	stmt := "UPDATE details SET key = ?, value = ? WHERE id = ?;"
	_, err := d.db.Exec(stmt, key, value, detail.ID)
	if err != nil {
		return nil, err
	}
	detail.Key = key
	detail.Value = value
	return &detail, nil
}

func (d *Details) Delete(detail Detail) error {
	stmt := "DELETE FROM details WHERE id = ?;"
	_, err := d.db.Exec(stmt, detail.ID)
	return err
}

func (d *Details) DeleteAll(tx *sql.Tx, profileID int) error {
	stmt := "DELETE FROM details WHERE profile_id = ?;"
	_, err := tx.Exec(stmt, profileID)
	return err
}
