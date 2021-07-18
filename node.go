package main

import (
	"database/sql"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PGxPoolNode struct {
	conn    *pgxpool.Pool
	primary bool
	health  bool
}

type SQLPoolNode struct {
	conn    *sql.DB
	primary bool
	health  bool
}
