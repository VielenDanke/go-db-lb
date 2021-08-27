package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/jackc/pgx/v4"

	"github.com/stretchr/testify/assert"
)

func TestLoadBalancer_CallFirstAvailable_PGxPoolNode(t *testing.T) {
	counter := 0

	for counter < 10 {
		pool := pgxLB.CallFirstAvailable().PGxPool()
		if pool == nil {
			t.Fatal(fmt.Errorf("node is nil"))
		}
		if fErr := pool.Ping(context.Background()); fErr != nil {
			t.Fatal(fErr)
		}
		counter++
	}
}

func TestLoadBalancer_CallFirstAvailable_Insert_PGxPoolNode(t *testing.T) {
	counter := 0

	stopCh := make(chan struct{})
	errCh := make(chan error, 0)
	wg := &sync.WaitGroup{}
	ctx := context.Background()

	for counter < 100 {
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, counter int) {
			pool := pgxLB.CallFirstAvailable().PGxPool()
			if pool == nil {
				errCh <- fmt.Errorf("node is nil")
				return
			}
			if fErr := pool.Ping(ctx); fErr != nil {
				errCh <- fErr
				return
			}
			tx, txErr := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
			if txErr != nil {
				errCh <- txErr
				return
			}
			exec, execErr := tx.Exec(ctx, "insert into users(name) values($1)", fmt.Sprintf("Test %d", counter))
			if execErr != nil {
				rollbackErr := tx.Rollback(ctx)
				if rollbackErr != nil {
					log.Println(rollbackErr)
				}
				errCh <- execErr
				return
			}
			if commitErr := tx.Commit(ctx); commitErr != nil {
				log.Println(commitErr)
			}
			assert.NotEqual(t, 0, exec.RowsAffected())
			wg.Done()
		}(ctx, wg, errCh, counter)
		counter++
	}
	go func(wg *sync.WaitGroup, stopCh chan struct{}) {
		wg.Wait()
		close(stopCh)
	}(wg, stopCh)

	select {
	case <-stopCh:
		log.Println("Test succeed")
	case err := <-errCh:
		t.Fatal(err)
	}
}

func TestLoadBalancer_CallPrimaryNode_PGxPoolNode(t *testing.T) {
	counter := 0

	for counter < 10 {
		fAvailable := pgxLB.CallPrimary().PGxPool()
		if fAvailable == nil {
			t.Error(fmt.Errorf("node is nil"))
			return
		}
		if fErr := fAvailable.Ping(context.Background()); fErr != nil {
			t.Error(fErr)
			return
		}
		counter++
	}
}

func TestLoadBalancer_CallPrimaryPreferred_PGxPoolNode(t *testing.T) {
	counter := 0

	for counter < 10 {
		fAvailable := pgxLB.CallPrimaryPreferred().PGxPool()
		if fAvailable == nil {
			t.Error(fmt.Errorf("node is nil"))
			return
		}
		if fErr := fAvailable.Ping(context.Background()); fErr != nil {
			t.Error(fErr)
			return
		}
		counter++
	}
}
