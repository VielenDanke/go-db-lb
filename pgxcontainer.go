package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jmoiron/sqlx"
	"time"
)

const primaryNode = 0

type loadBalancer struct {
	nodes []PoolNode
}

func NewLoadBalancer(ctx context.Context, nodeSize, timeoutSeconds int) *loadBalancer {
	nodes := make([]PoolNode, 0, nodeSize)
	lb := &loadBalancer{nodes: nodes}
	go func(lb *loadBalancer) {
		counter := 0
		for {
			time.Sleep(time.Duration(timeoutSeconds) * time.Second)
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
	}(lb)
	return lb
}

func (lb *loadBalancer) AddPGxPoolNode(ctx context.Context, n *pgxpool.Pool) error {
	if pingErr := n.Ping(ctx); pingErr != nil {
		return pingErr
	}
	lb.nodes = append(lb.nodes, &PGxPoolNode{conn: n, health: true})
	return nil
}

func (lb *loadBalancer) AddPGxPoolPrimaryNode(ctx context.Context, n *pgxpool.Pool) error {
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
		lb.nodes[primaryNode] = pn
	}
	return nil
}

func (lb *loadBalancer) AddSQLxNode(ctx context.Context, n *sqlx.DB) error {
	if pingErr := n.PingContext(ctx); pingErr != nil {
		return pingErr
	}
	lb.nodes = append(lb.nodes, &SQLxPoolNode{conn: n, health: true})
	return nil
}

func (lb *loadBalancer) AddSQLxPrimaryNode(ctx context.Context, n *sqlx.DB) error {
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
		lb.nodes[primaryNode] = pn
	}
	return nil
}

func (lb *loadBalancer) AddSQLNode(ctx context.Context, n *sql.DB) error {
	if pingErr := n.PingContext(ctx); pingErr != nil {
		return pingErr
	}
	lb.nodes = append(lb.nodes, &SQLPoolNode{conn: n, health: true})
	return nil
}

func (lb *loadBalancer) AddSQLPrimaryNode(ctx context.Context, n *sql.DB) error {
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
		lb.nodes[primaryNode] = pn
	}
	return nil
}

func (lb *loadBalancer) CallPrimaryPreferred() PoolNode {
	node := lb.CallPrimaryNode()
	if node == nil {
		return lb.CallFirstAvailable()
	}
	return node
}

func (lb *loadBalancer) CallPrimaryNode() PoolNode {
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

func (lb *loadBalancer) CallFirstAvailable() PoolNode {
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
	case <-time.Tick(2 * time.Second):
		return nil
	}
}

func (lb *loadBalancer) swapPGxPoolNode(n *PGxPoolNode) {
	temp := lb.nodes[0]
	lb.nodes[0] = n
	lb.nodes = append(lb.nodes, temp)
}

func (lb *loadBalancer) swapSQLxPoolNode(n *SQLxPoolNode) {
	temp := lb.nodes[0]
	lb.nodes[0] = n
	lb.nodes = append(lb.nodes, temp)
}

func (lb *loadBalancer) swapSQLPoolNode(n *SQLPoolNode) {
	temp := lb.nodes[0]
	lb.nodes[0] = n
	lb.nodes = append(lb.nodes, temp)
}
