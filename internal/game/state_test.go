package game

import "testing"

func TestStateStringForGUIReceipt(t *testing.T) {
	if got := StateLoadGame.String(); got != "StateLoadGame" {
		t.Fatalf("StateLoadGame.String() = %q", got)
	}
	if got := State(-1).String(); got != "State(-1)" {
		t.Fatalf("State(-1).String() = %q", got)
	}
}
