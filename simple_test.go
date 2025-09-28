package main

import "testing"

func TestSimple(t *testing.T) {
	if 1 != 1 {
		t.Error("1 should equal 1")
	}
}

func TestAddition(t *testing.T) {
	result := 2 + 2
	if result != 4 {
		t.Errorf("Expected 4, got %d", result)
	}
}
