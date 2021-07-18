package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadBalancer_CallFirstAvailable(t *testing.T) {
	counter := 0

	for counter < 10 {
		fAvailable, fErr := lb.CallFirstAvailable()
		if fErr != nil {
			t.Error(fErr)
		}
		if fErr = fAvailable.Ping(context.Background()); fErr != nil {
			t.Error(fErr)
		}
		counter++
	}
}

func TestLoadBalancer_CallFirstAvailable_Insert(t *testing.T) {
	counter := 0

	stopCh := make(chan struct{})
	wg := &sync.WaitGroup{}
	ctx := context.Background()

	for counter < 100 {
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup, t *testing.T, counter int) {
			fAvailable, fErr := lb.CallFirstAvailable()
			if fErr != nil {
				t.Error(fErr)
			}
			if fErr = fAvailable.Ping(ctx); fErr != nil {
				t.Error(fErr)
			}
			tx, txErr := fAvailable.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
			if txErr != nil {
				t.Error(txErr)
			}
			exec, execErr := tx.Exec(ctx, "insert into users(name) values($1)", fmt.Sprintf("Test %d", counter))
			if execErr != nil {
				rollbackErr := tx.Rollback(ctx)
				if rollbackErr != nil {
					log.Println(rollbackErr)
				}
				t.Error(execErr)
			}
			if commitErr := tx.Commit(ctx); commitErr != nil {
				log.Println(commitErr)
			}
			assert.NotEqual(t, 0, exec.RowsAffected())
			wg.Done()
		}(ctx, wg, t, counter)
		counter++
	}
	go func(wg *sync.WaitGroup, stopCh chan struct{}) {
		wg.Wait()
		close(stopCh)
	}(wg, stopCh)

	<-stopCh
}

func TestLoadBalancer_CallPrimaryNode(t *testing.T) {
	counter := 0

	for counter < 10 {
		fAvailable, fErr := lb.CallPrimaryNode()
		if fErr != nil {
			t.Error(fErr)
		}
		if fErr = fAvailable.Ping(context.Background()); fErr != nil {
			t.Error(fErr)
		}
		counter++
	}
}

func TestLoadBalancer_CallPrimaryPreferred(t *testing.T) {
	counter := 0

	for counter < 10 {
		fAvailable, fErr := lb.CallPrimaryPreferred()
		if fErr != nil {
			t.Error(fErr)
		}
		if fErr = fAvailable.Ping(context.Background()); fErr != nil {
			t.Error(fErr)
		}
		counter++
	}
}