package data

import (
	"database/sql"
	"errors"
)

type ProfileRepository struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (pr *ProfileRepository) GetDetailsByProfileName(profileName string) ([]Detail, error) {
	stmt := "select details.id, details.key, details.value, details.type, details.profile_id from details join profiles where details.profile_id = profiles.id and profiles.name = ?;"
	rows, err := pr.db.Query(stmt, profileName)
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

func (pr *ProfileRepository) GetAllDetails(profileId int) ([]Detail, error) {
	stmt := "SELECT id, key, value, type, profile_id FROM details WHERE profile_id = ?;"
	rows, err := pr.db.Query(stmt, profileId)
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

func (pr *ProfileRepository) AddDetail(key string, value string, detailType DetailType, profileID int) (*Detail, error) {
	var detail Detail
	stmt := "INSERT INTO details (key, value, type, profile_id) VALUES (?, ?, ?, ?);"
	res, err := pr.db.Exec(stmt, key, value, detailType.String(), profileID)
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

func (pr *ProfileRepository) UpdateDetail(detail Detail, key string, value string) (*Detail, error) {
	stmt := "UPDATE details SET key = ?, value = ? WHERE id = ?;"
	_, err := pr.db.Exec(stmt, key, value, detail.ID)
	if err != nil {
		return nil, err
	}
	detail.Key = key
	detail.Value = value
	return &detail, nil
}

func (pr *ProfileRepository) DeleteDetail(detail Detail) error {
	stmt := "DELETE FROM details WHERE id = ?;"
	_, err := pr.db.Exec(stmt, detail.ID)
	return err
}

func (pr *ProfileRepository) GetAllProfiles() ([]Profile, error) {
	stmt := "SELECT id, name from profiles"
	rows, err := pr.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	profiles := []Profile{}
	for rows.Next() {
		profile := &Profile{}
		err = rows.Scan(&profile.ID, &profile.Name)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, *profile)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return profiles, nil
}

func (pr *ProfileRepository) AddProfile(name string) (Profile, error) {
	var profile Profile
	stmt := "INSERT INTO profiles (name) VALUES (?);"
	res, err := pr.db.Exec(stmt, name)
	if err != nil {
		return profile, err
	}
	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return profile, err
	}
	return Profile{ID: int(id), Name: name}, nil
}

func (pr *ProfileRepository) UpdateProfile(profile Profile, newName string) (Profile, error) {
	stmt := "UPDATE profiles SET name = ? WHERE id = ?;"
	_, err := pr.db.Exec(stmt, newName, profile.ID)
	if err != nil {
		return profile, err
	}
	profile.Name = newName
	return profile, nil
}

func (pr *ProfileRepository) DeleteProfile(profile Profile) error {
	tx, err := pr.db.Begin()
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}
		return err
	}

	stmt := "DELETE FROM details WHERE profile_id = ?;"
	_, err = tx.Exec(stmt, profile.ID)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}
		return err
	}

	stmt = "DELETE FROM profiles WHERE id = ?;"
	_, err = tx.Exec(stmt, profile.ID)

	err = tx.Commit()
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}
		return err
	}
	return nil
}
