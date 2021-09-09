package container

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jmoiron/sqlx"
)

const primaryNode = 0

type LoadBalancer struct {
	nodes []PoolNode
}

func NewLoadBalancer(ctx context.Context, nodeSize, timeoutSeconds int) (*LoadBalancer, error) {
	if timeoutSeconds <= 0 {
		return nil, errors.New("timeout cannot be equal or less than 0")
	}
	nodes := make([]PoolNode, 0, nodeSize)
	ticker := time.NewTicker(time.Duration(timeoutSeconds) * time.Second)
	lb := &LoadBalancer{nodes: nodes}
	go func(lb *LoadBalancer, ch <-chan time.Time) {
		counter := 0
		for range ch {
			if counter > len(lb.nodes)-1 {
				counter = 0
			}
			node := lb.nodes[counter]
			if node == nil {
				continue
			}
			if pingErr := node.Ping(ctx); pingErr != nil {
				node.SetHealthyStatus(false)
				lb.nodes[counter] = node
			} else {
				node.SetHealthyStatus(true)
				lb.nodes[counter] = node
			}
			counter++
		}
	}(lb, ticker.C)
	return lb, nil
}

func (lb *LoadBalancer) AddPGxPoolNode(ctx context.Context, n *pgxpool.Pool) error {
	if err := checkForNil(ctx, n); err != nil {
		return err
	}
	if pingErr := n.Ping(ctx); pingErr != nil {
		return pingErr
	}
	lb.nodes = append(lb.nodes, &PGxPoolNode{conn: n, health: true})
	return nil
}

func (lb *LoadBalancer) AddPGxPoolPrimaryNode(ctx context.Context, n *pgxpool.Pool) error {
	if err := checkForNil(ctx, n); err != nil {
		return err
	}
	if pingErr := n.Ping(ctx); pingErr != nil {
		return pingErr
	}
	pn := &PGxPoolNode{conn: n, primary: true, health: true}
	if len(lb.nodes) > 0 {
		if lb.nodes[0].IsPrimary() {
			return fmt.Errorf("primary node already exists")
		}
		lb.swapPGxPoolNode(pn)
	} else {
		lb.nodes = append(lb.nodes, pn)
	}
	return nil
}

func (lb *LoadBalancer) AddSQLxNode(ctx context.Context, n *sqlx.DB) error {
	if err := checkForNil(ctx, n); err != nil {
		return err
	}
	if pingErr := n.PingContext(ctx); pingErr != nil {
		return pingErr
	}
	lb.nodes = append(lb.nodes, &SQLxPoolNode{conn: n, health: true})
	return nil
}

func (lb *LoadBalancer) AddSQLxPrimaryNode(ctx context.Context, n *sqlx.DB) error {
	if err := checkForNil(ctx, n); err != nil {
		return err
	}
	if pingErr := n.PingContext(ctx); pingErr != nil {
		return pingErr
	}
	pn := &SQLxPoolNode{conn: n, primary: true, health: true}
	if len(lb.nodes) > 0 {
		if lb.nodes[0].IsPrimary() {
			return fmt.Errorf("primary node already exists")
		}
		lb.swapSQLxPoolNode(pn)
	} else {
		lb.nodes = append(lb.nodes, pn)
	}
	return nil
}

func (lb *LoadBalancer) AddSQLNode(ctx context.Context, n *sql.DB) error {
	if err := checkForNil(ctx, n); err != nil {
		return err
	}
	if pingErr := n.PingContext(ctx); pingErr != nil {
		return pingErr
	}
	lb.nodes = append(lb.nodes, &SQLPoolNode{conn: n, health: true})
	return nil
}

func (lb *LoadBalancer) AddSQLPrimaryNode(ctx context.Context, n *sql.DB) error {
	if err := checkForNil(ctx, n); err != nil {
		return err
	}
	if pingErr := n.PingContext(ctx); pingErr != nil {
		return pingErr
	}
	pn := &SQLPoolNode{conn: n, primary: true, health: true}
	if len(lb.nodes) > 0 {
		if lb.nodes[0].IsPrimary() {
			return fmt.Errorf("primary node already exists")
		}
		lb.swapSQLPoolNode(pn)
	} else {
		lb.nodes = append(lb.nodes, pn)
	}
	return nil
}

func (lb *LoadBalancer) CallPrimaryPreferred() PoolNode {
	node := lb.CallPrimary()
	if node == nil {
		return lb.CallFirstAvailable()
	}
	return node
}

func (lb *LoadBalancer) CallPrimary() PoolNode {
	pr := lb.nodes[primaryNode]
	if pr == nil {
		return nil
	}
	if !pr.IsPrimary() {
		return nil
	}
	if !pr.IsHealthy() {
		return nil
	}
	return pr
}

func (lb *LoadBalancer) CallFirstAvailable() PoolNode {
	nCh := make(chan PoolNode, 1)
	for _, v := range lb.nodes {
		go func(v PoolNode, nCh chan PoolNode) {
			if v.IsHealthy() {
				nCh <- v
			} else {
				return
			}
		}(v, nCh)
	}
	select {
	case conn := <-nCh:
		return conn
	case <-time.NewTicker(2 * time.Second).C:
		return nil
	}
}

func (lb *LoadBalancer) swapPGxPoolNode(n *PGxPoolNode) {
	temp := lb.nodes[0]
	lb.nodes[0] = n
	lb.nodes = append(lb.nodes, temp)
}

func (lb *LoadBalancer) swapSQLxPoolNode(n *SQLxPoolNode) {
	temp := lb.nodes[0]
	lb.nodes[0] = n
	lb.nodes = append(lb.nodes, temp)
}

func (lb *LoadBalancer) swapSQLPoolNode(n *SQLPoolNode) {
	temp := lb.nodes[0]
	lb.nodes[0] = n
	lb.nodes = append(lb.nodes, temp)
}

func checkForNil(ctx context.Context, n interface{}) error {
	if n == nil {
		return errors.New("pool cannot be nil")
	}
	if ctx == nil {
		return errors.New("context cannot be nil")
	}
	return nil
}
