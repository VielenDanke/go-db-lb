package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"os"
	"testing"
)

var pgxLB *loadBalancer

func init() {
	dbUrl := "user=user password=password sslmode=disable dbname=user host=localhost port=5432"
	ctx := context.Background()
	pgxLB = NewLoadBalancer(ctx, 2, 2)
	cfg, _ := pgxpool.ParseConfig(dbUrl)
	cfg.MaxConns = 10
	cfg.MinConns = 5
	cfg.MaxConnLifetime = 60
	poolF, _ := pgxpool.ConnectConfig(ctx, cfg)
	poolS, _ := pgxpool.ConnectConfig(ctx, cfg)
	pgxLB.AddPGxPoolNode(ctx, poolF)
	pgxLB.AddPGxPoolPrimaryNode(ctx, poolS)
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
