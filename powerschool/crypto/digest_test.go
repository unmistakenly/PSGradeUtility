package crypto

import "testing"

func TestDigestResponse_KnownVector(t *testing.T) {
	// Expected value computed independently (Python hashlib.md5, not derived
	// from this Go code) for the fixed inputs below — a real regression
	// vector, not a tautology against the implementation under test.
	got := DigestResponse("testnonce123", "testcnonce456")
	want := "ee66a8bb48999f7f26e2cf8704570215"
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestDigestResponse_Deterministic(t *testing.T) {
	a := DigestResponse("n1", "c1")
	b := DigestResponse("n1", "c1")
	if a != b {
		t.Fatalf("same inputs produced different output: %s vs %s", a, b)
	}
}

func TestDigestResponse_DifferentNonceDifferentOutput(t *testing.T) {
	a := DigestResponse("n1", "c1")
	b := DigestResponse("n2", "c1")
	if a == b {
		t.Fatal("different nonces produced the same digest")
	}
}
