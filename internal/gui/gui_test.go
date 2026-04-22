package gui

import "testing"

func TestNewPreservesCheckoutMode(t *testing.T) {
	gui, err := New(string(CheckoutMode), nil)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if got := gui.State.Mode.ModeID; got != CheckoutMode {
		t.Fatalf("expected mode %q, got %q", CheckoutMode, got)
	}
}
