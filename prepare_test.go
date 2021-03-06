package container

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var pgxLB *LoadBalancer
var sqlLB *LoadBalancer
var sqlxLB *LoadBalancer

func init() {
	dbUrl := "user=user password=password sslmode=disable dbname=test host=localhost port=5432"
	ctx := context.Background()
	pgxLB, _ = NewLoadBalancer(ctx, 2, 2)
	sqlLB, _ = NewLoadBalancer(ctx, 2, 2)
	sqlxLB, _ = NewLoadBalancer(ctx, 2, 2)
	cfg, _ := pgxpool.ParseConfig(dbUrl)
	cfg.MaxConns = 10
	cfg.MinConns = 5
	cfg.MaxConnLifetime = 60
	poolF, _ := pgxpool.ConnectConfig(ctx, cfg)
	poolS, _ := pgxpool.ConnectConfig(ctx, cfg)
	dbF, _ := sql.Open("postgres", dbUrl)
	dbS, _ := sql.Open("postgres", dbUrl)
	dbxF, _ := sqlx.Open("postgres", dbUrl)
	dbxS, _ := sqlx.Open("postgres", dbUrl)
	pgxLB.AddPGxPoolNode(ctx, poolF)
	pgxLB.AddPGxPoolPrimaryNode(ctx, poolS)
	sqlLB.AddSQLPrimaryNode(ctx, dbF)
	sqlLB.AddSQLNode(ctx, dbS)
	sqlxLB.AddSQLxNode(ctx, dbxS)
	sqlxLB.AddSQLxPrimaryNode(ctx, dbxF)
}

func TestMain(m *testing.M) {
	code := m.Run()
	_, delErr := pgxLB.CallPrimaryPreferred().PGxPool().Exec(context.Background(), "delete from users")
	if delErr != nil {
		log.Fatalln(delErr)
	}
	os.Exit(code)
}
