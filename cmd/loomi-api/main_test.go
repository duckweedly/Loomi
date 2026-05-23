package main

import "testing"

func TestProductServiceForPoolReturnsNilWhenPoolUnavailable(t *testing.T) {
	if service := productServiceForPool(nil); service != nil {
		t.Fatalf("service = %#v, want nil", service)
	}
}
