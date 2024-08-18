package data

import "database/sql"

type Profile struct {
    ID int
    Name string
}

type Profiles struct {
    db *sql.DB
}

func (p *Profiles) GetAll() ([]Profile, error) {
    return []Profile{}, nil
}

func (p *Profiles) Add(name string) (Profile, error) {
    return Profile{}, nil
}

func (p *Profiles) Update(profile Profile, newName string) error {
    return nil
}

func (p *Profiles) Delete(name string) error {
    return nil
}

