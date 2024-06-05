package data

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrItemDoesNotExist = errors.New("insert or update on table \"orders\" violates foreign key constraint \"orders_item_id_fkey\"")
	ErrUserDoesNotExist = errors.New("insert or update on table \"orders\" violates foreign key constraint \"orders_user_id_fkey\"")
)

func New(dsn string) (*OrderStorage, error) {
	const op = "data.sqlite.New"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &OrderStorage{DB: db}, nil
}
