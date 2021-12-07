package main

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"time"

	"go.uber.org/multierr"

	_ "github.com/lib/pq"
)

var (
	//go:embed sql/schema.sql
	schemaSQL string

	//go:embed sql/add.sql
	addSQL string

	//go:embed sql/query.sql
	querySQL string

	//go:embed sql/last.sql
	lastSQL string
)

type DB struct {
	db *sql.DB
}

func NewDB(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(schemaSQL); err != nil {
		return nil, multierr.Append(err, db.Close())
	}

	return &DB{db}, nil
}

func (d *DB) Add(ctx context.Context, entry Entry) error {
	log.Print(addSQL)
	_, err := d.db.ExecContext(ctx, addSQL, entry.Time, entry.User, entry.Content)
	/* FIXME:
		sql.Named("time", entry.Time),
		sql.Named("login", entry.User),
		sql.Named("content", entry.Content),
	)
	*/
	return err
}

func (d *DB) Query(ctx context.Context, start, end time.Time) ([]Entry, error) {
	rows, err := d.db.QueryContext(ctx, querySQL, start, end)
	/* FIXME:
		sql.Named("start", start),
		sql.Named("end", end),
	)
	*/
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.Time, &e.User, &e.Content); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func (d *DB) Last(ctx context.Context) (Entry, error) {
	row := d.db.QueryRowContext(ctx, lastSQL)
	var e Entry
	if err := row.Scan(&e.Time, &e.User, &e.Content); err != nil {
		return Entry{}, err
	}

	return e, nil
}

func (d *DB) Health(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *DB) Close() error {
	return d.db.Close()
}
