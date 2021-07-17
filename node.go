package main

import "github.com/jackc/pgx/v4/pgxpool"

type PoolNode struct {
	conn    *pgxpool.Pool
	primary bool
	health  bool
}
