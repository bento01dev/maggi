package data

import "database/sql"

type Profile struct {
    ID int
    Name string
}

type Profiles struct {
    db *sql.DB
}
