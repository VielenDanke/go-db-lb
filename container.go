package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type loadBalancer struct {
	nodes []*PoolNode
}

func NewLoadBalancer(ctx context.Context, nodeSize, timeoutSeconds int) *loadBalancer {
	nodes := make([]*PoolNode, 0, nodeSize)
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
			if pingErr := node.conn.Ping(ctx); pingErr != nil {
				node.health = false
				lb.nodes[counter] = node
			} else {
				node.health = true
				lb.nodes[counter] = node
			}
			counter++
		}
	}(lb)
	return lb
}

func (lb *loadBalancer) AddNode(ctx context.Context, n *pgxpool.Pool) error {
	if pingErr := n.Ping(ctx); pingErr != nil {
		return pingErr
	}
	lb.nodes = append(lb.nodes, &PoolNode{conn: n, health: true})
	return nil
}

func (lb *loadBalancer) AddPrimaryNode(ctx context.Context, n *pgxpool.Pool) error {
	if pingErr := n.Ping(ctx); pingErr != nil {
		return pingErr
	}
	pn := &PoolNode{conn: n, primary: true, health: true}
	if len(lb.nodes) > 0 {
		if lb.nodes[0].primary {
			return fmt.Errorf("primary node already exists")
		}
		lb.swap(pn)
	} else {
		lb.nodes[0] = pn
	}
	return nil
}

func (lb *loadBalancer) CallPrimaryPreferred() (*pgxpool.Pool, error) {
	node, err := lb.CallPrimaryNode()
	if err != nil {
		return lb.CallFirstAvailable()
	}
	return node, nil
}

func (lb *loadBalancer) CallPrimaryNode() (*pgxpool.Pool, error) {
	pr := lb.nodes[0]
	if pr == nil {
		return nil, fmt.Errorf("lb nodes are empty")
	}
	if !pr.primary {
		return nil, fmt.Errorf("primary node not found")
	}
	return pr.conn, nil
}

func (lb *loadBalancer) CallFirstAvailable() (*pgxpool.Pool, error) {
	nCh := make(chan *pgxpool.Pool, 1)
	for _, v := range lb.nodes {
		go func(v *PoolNode, nCh chan *pgxpool.Pool) {
			if v.health {
				nCh <- v.conn
			} else {
				return
			}
		}(v, nCh)
	}
	select {
	case conn := <-nCh:
		return conn, nil
	case <-time.Tick(2 * time.Second):
		return nil, fmt.Errorf("no available nodes found")
	}
}

func (lb *loadBalancer) swap(n *PoolNode) {
	temp := lb.nodes[0]
	lb.nodes[0] = n
	lb.nodes = append(lb.nodes, temp)
}
