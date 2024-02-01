package csvutil

import (
	"fmt"

	"gorm.io/gorm"
)

type Cursor[T any] struct {
	DBconn    *gorm.DB
	Query     string
	FetchSize int
}

func NewCursor[T any](dbconn *gorm.DB, query string, fetchSize int) (*Cursor[T], error) {
	cursor := &Cursor[T]{
		DBconn:    dbconn,
		Query:     query,
		FetchSize: fetchSize,
	}

	// Start a transaction
	cursor.DBconn = cursor.DBconn.Begin()
	// Declare a cursor
	if err := cursor.DBconn.Exec("DECLARE mycursor CURSOR FOR " + query).Error; err != nil {
		return nil, err
	}

	return cursor, nil
}

func (cursor *Cursor[T]) FetchCursor() ([]T, error) {
	holder := make([]T, 0)
	stmt := fmt.Sprintf("FETCH %d FROM mycursor", cursor.FetchSize)
	// Fetch the results
	if err := cursor.DBconn.Raw(stmt).Scan(&holder).Error; err != nil {
		return nil, err
	}
	return holder, nil
}

func (cursor *Cursor[T]) Close() error {
	// Close the cursor
	err := cursor.DBconn.Exec("CLOSE mycursor").Error
	if err != nil {
		return err
	}

	// Commit the transaction
	err = cursor.DBconn.Commit().Error
	if err != nil {
		return err
	}

	return nil
}
