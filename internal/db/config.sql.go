// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: config.sql

package db

import (
	"context"
)

const getPollingInterval = `-- name: GetPollingInterval :one
select time from config limit 1
`

func (q *Queries) GetPollingInterval(ctx context.Context) (int32, error) {
	row := q.db.QueryRow(ctx, getPollingInterval)
	var time int32
	err := row.Scan(&time)
	return time, err
}
