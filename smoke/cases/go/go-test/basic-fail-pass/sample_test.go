package smoke

import "testing"

func TestPass(t *testing.T) {}

func TestFail(t *testing.T) {
	t.Fatalf("intentional failure")
}
