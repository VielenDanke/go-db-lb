package main

import (
	"context"
	"testing"
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