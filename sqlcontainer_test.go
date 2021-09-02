package container

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadBalancer_CallFirstAvailable_SQLPoolNode(t *testing.T) {
	counter := 0

	for counter < 10 {
		dbx := sqlLB.CallFirstAvailable().DB()
		if dbx == nil {
			t.Fatal(fmt.Errorf("node is nil"))
		}
		if fErr := dbx.PingContext(context.Background()); fErr != nil {
			t.Fatal(fErr)
		}
		counter++
	}
}

func TestLoadBalancer_CallFirstAvailable_Insert_SQLPoolNode(t *testing.T) {
	counter := 0

	stopCh := make(chan struct{})
	errCh := make(chan error)
	wg := &sync.WaitGroup{}
	ctx := context.Background()

	for counter < 100 {
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, counter int) {
			pool := sqlLB.CallFirstAvailable().DB()
			if pool == nil {
				errCh <- fmt.Errorf("node is nil")
				return
			}
			if fErr := pool.PingContext(ctx); fErr != nil {
				errCh <- fErr
				return
			}
			tx, txErr := pool.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
			if txErr != nil {
				errCh <- txErr
				return
			}
			exec, execErr := tx.ExecContext(ctx, "insert into users(name) values($1)", fmt.Sprintf("Test %d", counter))
			if execErr != nil {
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					log.Println(rollbackErr)
				}
				errCh <- execErr
				return
			}
			if commitErr := tx.Commit(); commitErr != nil {
				log.Println(commitErr)
			}
			affRows, affErr := exec.RowsAffected()
			assert.Nil(t, affErr)
			assert.NotEqual(t, 0, affRows)
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

func TestLoadBalancer_CallPrimaryNode_SQLPoolNode(t *testing.T) {
	counter := 0

	for counter < 10 {
		fAvailable := sqlLB.CallPrimary().DB()
		if fAvailable == nil {
			t.Error(fmt.Errorf("node is nil"))
			return
		}
		if fErr := fAvailable.PingContext(context.Background()); fErr != nil {
			t.Error(fErr)
			return
		}
		counter++
	}
}

func TestLoadBalancer_CallPrimaryPreferred_SQLPoolNode(t *testing.T) {
	counter := 0

	for counter < 10 {
		fAvailable := sqlLB.CallPrimaryPreferred().DB()
		if fAvailable == nil {
			t.Error(fmt.Errorf("node is nil"))
			return
		}
		if fErr := fAvailable.PingContext(context.Background()); fErr != nil {
			t.Error(fErr)
			return
		}
		counter++
	}
}
