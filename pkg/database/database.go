// Package database provides support for access the database.
package database

import (
	"context"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	// healthCheckPeriod is the duration interval between health
	// checks.
	healthCheckPeriod = 5 * time.Second

	// maxConnIdletime is the maximum amount of time after which an
	// idle connection will be automatically closed.
	maxConnIdletime = 10 * time.Second

	// maxConnLifetime is the maximum amount of time for a connection.
	maxConnLifetime = 30 * time.Second

	// maxConns is the maximum number of connections for a database
	// pool.
	maxConns int32 = 10

	// minConns is the maximum number of connections for a database
	// pool.
	minConns int32 = 5
)

// Open returns a database connection.
func Open(ctx context.Context, connString string) (*Client, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	cfg.HealthCheckPeriod = healthCheckPeriod
	cfg.MaxConnIdleTime = maxConnIdletime
	cfg.MaxConnLifetime = maxConnLifetime
	cfg.MaxConns = maxConns
	cfg.MinConns = minConns

	pool, err := pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	client := NewClient(pool)
	err = Test(ctx, client)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Test returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func Test(ctx context.Context, client livenessClient) error {
	// Ping the database.
	var pingError error
	for attempts := 1; ; attempts++ {
		pingError = client.Ping(ctx)
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	// Make sure we didn't timeout or get cancelled.
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Run a simple query to determine connectivity. Running this
	// query forces a round trip through the database.
	return client.Test(ctx)
}

type livenessClient interface {
	Ping(context.Context) error
	Test(context.Context) error
}

// Client wraps the connection pool to provide logging.
type Client struct {
	db          *pgxpool.Pool
	serviceName string
}

func NewClient(db *pgxpool.Pool) *Client {
	return &Client{
		db:          db,
		serviceName: db.Config().ConnConfig.Database,
	}
}

// Close closes the underlying connection pool.
func (c *Client) Close() {
	c.db.Close()
}

func (c *Client) Test(ctx context.Context) error {
	const q = `SELECT true`
	var tmp bool
	return c.db.QueryRow(ctx, q).Scan(&tmp)
}

func (c *Client) Ping(ctx context.Context) error {
	return c.db.Ping(ctx)
}

func (c *Client) DB() *pgxpool.Pool {
	return c.db
}

// Exec executes the provided query as a prepared statement.
func (c *Client) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	tag, err := c.db.Exec(ctx, sql, args...)
	return tag, err
}

// QueryRow executes the provided query as a prepared statement.
func (c *Client) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return c.db.QueryRow(ctx, sql, args...)
}

// Query executes the provided query as a prepared statement.
func (c *Client) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	rows, err := c.db.Query(ctx, sql, args...)
	return rows, err
}

// Transact executes the provided query within a transaction. Rollbacks are
// triggered if any errors are returned or panic encountered.
//
// Example usage:
//
//	err := Transact(ctx, db, func(tx *Tx) error {
//		rows, terr := tx.exec(ctx, `UPDATE ads SET status = 'canceled' WHERE id = 1`)
//		if terr != nil {
//			return terr
//		}
//	}
func (c *Client) Transact(ctx context.Context, txFunc func(tx *Tx) error, txOpt pgx.TxOptions) (err error) {
	pgxtx, err := c.begin(ctx, txOpt)
	if err != nil {
		return err
	}

	tx := &Tx{pgxtx: pgxtx, serviceName: c.serviceName}
	defer func() {
		// If the transaction function panics, rollback and panic
		// again.
		if p := recover(); p != nil {
			_ = tx.pgxtx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			rerr := tx.pgxtx.Rollback(ctx) // err is non-nil; don't change it
			if rerr != nil {
				panic(rerr)
			}
		} else {
			err = tx.pgxtx.Commit(ctx)
		}
	}()
	return txFunc(tx)
}

// Tx wraps a pgx.Tx to provide logging.
type Tx struct {
	pgxtx       pgx.Tx
	serviceName string
}

// Exec executes the provided query as a prepared statement.
func (t *Tx) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	tag, err := t.pgxtx.Exec(ctx, sql, args...)
	return tag, err
}

// QueryRow executes the provided query as a prepared statement.
func (t *Tx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return t.pgxtx.QueryRow(ctx, sql, args...)
}

// Query executes the provided query as a prepared statement.
func (t *Tx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	rows, err := t.pgxtx.Query(ctx, sql, args...)
	return rows, err
}

func (t *Tx) Begin(ctx context.Context) (pgx.Tx, error) {
	return t.pgxtx.Begin(ctx)
}

func (t *Tx) BeginFunc(ctx context.Context, f func(pgx.Tx) error) error {
	return t.pgxtx.BeginFunc(ctx, f)
}

func (t *Tx) Commit(ctx context.Context) error {
	return t.pgxtx.Commit(ctx)
}

func (t *Tx) Rollback(ctx context.Context) error {
	return t.pgxtx.Rollback(ctx)
}

func (t *Tx) Conn() *pgx.Conn {
	return t.pgxtx.Conn()
}

func (t *Tx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return t.pgxtx.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (t *Tx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return t.pgxtx.SendBatch(ctx, b)
}

func (t *Tx) LargeObjects() pgx.LargeObjects {
	return t.pgxtx.LargeObjects()
}

func (t *Tx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return t.pgxtx.Prepare(ctx, name, sql)
}

func (t *Tx) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	return t.pgxtx.QueryFunc(ctx, sql, args, scans, f)
}

// begin returns a pgx.tx transaction.
func (c *Client) begin(ctx context.Context, txOpt pgx.TxOptions) (pgx.Tx, error) {
	tag, err := c.db.BeginTx(ctx, txOpt)
	return tag, err
}
