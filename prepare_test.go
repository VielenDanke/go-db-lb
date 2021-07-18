package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"os"
	"testing"
)

var lb *loadBalancer

func init() {
	dbUrl := "user=user password=password sslmode=disable dbname=user host=localhost port=5432"
	ctx := context.Background()
	lb = NewPGxLoadBalancer(ctx, 2, 2)
	cfg, _ := pgxpool.ParseConfig(dbUrl)
	cfg.MaxConns = 10
	cfg.MinConns = 5
	cfg.MaxConnLifetime = 60
	poolF, _ := pgxpool.ConnectConfig(ctx, cfg)
	poolS, _ := pgxpool.ConnectConfig(ctx, cfg)
	lb.AddNode(ctx, poolF)
	lb.AddPrimaryNode(ctx, poolS)
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
