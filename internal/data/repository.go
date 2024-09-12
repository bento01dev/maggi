package data

import "database/sql"

type ProfileRepository struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) ProfileRepository {
	return ProfileRepository{db: db}
}

func (pr ProfileRepository) GetDetailsByProfileName(profileName string) ([]Detail, error) {
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
