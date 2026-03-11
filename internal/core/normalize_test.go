package core

import "testing"

func TestNormalizeContent(t *testing.T) {
	input := "  ```Make Build```\n`./mo-service`  "
	got := NormalizeContent(input)
	want := "make build ./mo-service"
	if got != want {
		t.Fatalf("NormalizeContent() = %q, want %q", got, want)
	}
}
