package crypto_test

// External test package (not `package crypto`) specifically so this can
// import both powerschool and powerschool/crypto without recreating the
// import cycle client.go's move caused — crypto can't import powerschool
// itself (see nonce.go), so its date-format layout string is duplicated
// there rather than shared. This test is the drift guard: it fails the
// moment the two layout strings diverge.

import (
	"testing"

	"github.com/unmistakenly/PSGradeUtility/powerschool"
)

const cryptoDateFormat = "2006-01-02T15:04:05.000Z" // must match nonce.go's local copy

func TestDateFormatMatchesPowerschoolTimeFormat(t *testing.T) {
	if cryptoDateFormat != powerschool.TimeFormat {
		t.Fatalf("crypto's duplicated date format %q has drifted from powerschool.TimeFormat %q", cryptoDateFormat, powerschool.TimeFormat)
	}
}
