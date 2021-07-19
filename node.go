package main

import (
	"context"
	"database/sql"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jmoiron/sqlx"
)

type PoolNode interface {
	DBx() *sqlx.DB
	DB() *sql.DB
	PGxPool() *pgxpool.Pool
	Ping(ctx context.Context) error
	IsPrimary() bool
	IsHealthy() bool
	SetHealthyStatus(status bool)
	SetPrimaryStatus(status bool)
}

type PGxPoolNode struct {
	conn    *pgxpool.Pool
	primary bool
	health  bool
}

func (pgx *PGxPoolNode) DBx() *sqlx.DB {
	return nil
}

func (pgx *PGxPoolNode) DB() *sql.DB {
	return nil
}

func (pgx *PGxPoolNode) PGxPool() *pgxpool.Pool {
	return pgx.conn
}

func (pgx *PGxPoolNode) Ping(ctx context.Context) error {
	return pgx.conn.Ping(ctx)
}

func (pgx *PGxPoolNode) IsPrimary() bool {
	return pgx.primary
}

func (pgx *PGxPoolNode) IsHealthy() bool {
	return pgx.health
}

func (pgx *PGxPoolNode) SetHealthyStatus(status bool) {
	pgx.health = status
}

func (pgx *PGxPoolNode) SetPrimaryStatus(status bool) {
	pgx.primary = status
}

type SQLPoolNode struct {
	conn    *sql.DB
	primary bool
	health  bool
}

func (sql *SQLPoolNode) DBx() *sqlx.DB {
	return nil
}

func (sql *SQLPoolNode) DB() *sql.DB {
	return sql.conn
}

func (sql *SQLPoolNode) PGxPool() *pgxpool.Pool {
	return nil
}

func (sql *SQLPoolNode) Ping(ctx context.Context) error {
	return sql.conn.PingContext(ctx)
}

func (sql *SQLPoolNode) IsPrimary() bool {
	return sql.primary
}

func (sql *SQLPoolNode) IsHealthy() bool {
	return sql.health
}

func (sql *SQLPoolNode) SetHealthyStatus(status bool) {
	sql.health = status
}

func (sql *SQLPoolNode) SetPrimaryStatus(status bool) {
	sql.primary = status
}

type SQLxPoolNode struct {
	conn    *sqlx.DB
	primary bool
	health  bool
}

func (sqlx *SQLxPoolNode) DBx() *sqlx.DB {
	return sqlx.conn
}

func (sqlx *SQLxPoolNode) DB() *sql.DB {
	return nil
}

func (sqlx *SQLxPoolNode) PGxPool() *pgxpool.Pool {
	return nil
}

func (sqlx *SQLxPoolNode) Ping(ctx context.Context) error {
	return sqlx.conn.PingContext(ctx)
}

func (sqlx *SQLxPoolNode) IsPrimary() bool {
	return sqlx.primary
}

func (sqlx *SQLxPoolNode) IsHealthy() bool {
	return sqlx.health
}

func (sqlx *SQLxPoolNode) SetHealthyStatus(status bool) {
	sqlx.health = status
}

func (sqlx *SQLxPoolNode) SetPrimaryStatus(status bool) {
	sqlx.primary = status
}
