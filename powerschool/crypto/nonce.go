package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	"time"

	"github.com/unmistakenly/PSGradeUtility/powerschool"
)

// two types of nonces are used by powerschool:
// 1. 8 random bytes, in hex format and base64-encoded, used to get the `serviceTicket` with the user's credentials
// 2. 16 random bytes in hex format, used as `cnonce` when requesting user data from the /PublicPortalServiceJSON endpoint
func Nonce(bites int64, b64 bool) (nonce string, date string) {
	b := &bytes.Buffer{}

	func() {
		var w io.Writer
		if b64 {
			w = base64.NewEncoder(base64.StdEncoding, b)
			defer w.(io.Closer).Close()
		} else {
			w = b
		}

		he := hex.NewEncoder(w)
		if _, err := io.CopyN(he, rand.Reader, bites); err != nil {
			panic(err) // this really should never happen
		}
	}()

	return b.String(), time.Now().UTC().Format(powerschool.TimeFormat)
}
