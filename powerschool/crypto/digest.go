package crypto

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

func md5Sum(data string) string {
	md := md5.Sum([]byte(data))
	return hex.EncodeToString(md[:])
}

// im writing this at 4:33am right now.
// it took me that long to find out that this CRAP app is using some random default credentials.
// thanks. i love powerschool so much.
// https://en.wikipedia.org/wiki/Digest_access_authentication#Overview
func DigestResponse(nonce, cnonce string) string {
	p1 := md5Sum("pearson:Protected:m0bApP5")
	p2 := md5Sum("POST:/pearson-rest/services/PublicPortalServiceJSON?response=application/json")
	return md5Sum(fmt.Sprintf("%s:%s:00000001:%s:auth:%s", p1, nonce, cnonce, p2))
}
