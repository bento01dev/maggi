package data

import "database/sql"

type Profile struct {
	ID   int
	Name string
}

type Profiles struct {
	db *sql.DB
}

func (p *Profiles) GetAll() ([]Profile, error) {
	stmt := "SELECT id, name from profiles"
	rows, err := p.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	profiles := []Profile{}
	for rows.Next() {
		p := &Profile{}
		err = rows.Scan(&p.ID, &p.Name)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return profiles, nil
}

func (p *Profiles) Add(name string) (Profile, error) {
	var profile Profile
	stmt := "INSERT INTO profiles (name) VALUES (?);"
	res, err := p.db.Exec(stmt, name)
	if err != nil {
		return profile, err
	}
	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return profile, err
	}
	return Profile{ID: int(id), Name: name}, nil
}

func (p *Profiles) Update(profile Profile, newName string) (Profile, error) {
    stmt := "UPDATE profiles SET name = ? WHERE id = ?;"
    _, err := p.db.Exec(stmt, newName, profile.ID)
    if err != nil {
        return profile, err
    }
    profile.Name = newName
	return profile, nil
}

func (p *Profiles) Delete(profile Profile) error {
    stmt := "DELETE FROM profiles WHERE id = ?;"
    _, err := p.db.Exec(stmt, profile.ID)
    return err
}
