package balancing_strategies

import (
	"balancer/internal/handlers"
	"balancer/internal/models"
	"balancer/util/logger"
	"context"
	"errors"
)

type RoundRobin struct {
	next uint
}

func NewRoundRobinBalancer() *RoundRobin {
	return &RoundRobin{
		next: 0,
	}
}

func (r *RoundRobin) SelectNode(ctx context.Context, nodes []*models.Node) (*models.Node, error) {
	n := len(nodes)
	if n == 0 {
		return nil, errors.New("no nodes available")
	}
	node := nodes[r.next]
	r.next = (r.next + 1) % uint(n)

	logger.Infof(ctx, "Strategy '%s' selected node: %s", r.Name(), node.URL.String())

	return node, nil
}

func (r *RoundRobin) Name() string {
	return "round_robin"
}

func (r *RoundRobin) PrepareData() (worker handlers.WorkerFunc) {
	return nil
}
