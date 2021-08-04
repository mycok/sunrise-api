package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict = errors.New("edit conflict")
)

type Models struct {
	Movies MovieModel
}

func New(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}
