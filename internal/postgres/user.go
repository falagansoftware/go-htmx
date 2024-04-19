package postgres

import (
	"context"
	"strings"

	gohtmx "github.com/falagansoftware/go-htmx/internal"
)

type UserService struct {
	db *DB
}

func NewUserService(db *DB) *UserService {
	return &UserService{db: db}
}

func (u *UserService) FindUserById(ctx context.Context, id string) (*gohtmx.User, error) {
	tx, err := u.db.BeginTx(ctx, nil)

	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user,err := findUserById(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserService) FindUsers(ctx context.Context, filters gohtmx.UserFilters) ([]*gohtmx.User, error) {
	tx, err := u.db.BeginTx(ctx, nil)

	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	users, _, err := findUsers(ctx, tx, filters)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Helpers

func findUserById(ctx context.Context, tx *Tx, id string) (*gohtmx.User, error) {
	users, _, err := findUsers(ctx, tx, gohtmx.UserFilters{Id: &id})

	if err != nil {
		return nil, err
	} else if len(users) == 0 {
		return nil, &gohtmx.Error{Code: gohtmx.ENOTFOUND, Message: "User not found"}
	}
	return users[0], nil

}

func findUsers(ctx context.Context, tx *Tx, filter gohtmx.UserFilters) (u []*gohtmx.User, n int, e error) {
	// Where clause based on filters props
	where, args := []string{"1=1"}, []interface{}{}

	if v := filter.Id; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}

	if v := filter.Name; v != nil {
		where, args = append(where, "name = ?"), append(args, *v)
	}

	if v := filter.Surname; v != nil {
		where, args = append(where, "surname = ?"), append(args, *v)
	}

	if v := filter.Email; v != nil {
		where, args = append(where, "email = ?"), append(args, *v)
	}

	if v := filter.Active; v != nil {
		where, args = append(where, "active = ?"), append(args, *v)
	}

	// Execute query
	rows, err := tx.QueryContext(ctx, `
		SELECT 
			id, 
			name, 
			surname, 
			email, 
			active, 
			created_at, 
			updated_at 
		FROM users 
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY id ASC
		`+FormatLimitOffset(filter.Limit, filter.Offset), args...)

	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	// Map rows to struct

	users := make([]*gohtmx.User, 0)

	for rows.Next() {
		var user gohtmx.User
		err := rows.Scan(&user.Id, &user.Name, &user.Surname, &user.Email, &user.Active, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}

	// Check rows error
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, len(users), nil

}