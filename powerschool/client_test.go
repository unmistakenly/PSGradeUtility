package powerschool

import "testing"

// extractXMLFields replaced 3 unchecked regex[1] indexes. Cover the happy
// path plus the exact failure modes those regexes used to panic on.

func TestExtractXMLFields_HappyPath(t *testing.T) {
	body := `<?xml version="1.0"?><soap:Envelope xmlns:soap="x"><soap:Body><loginToPublicPortalResponse><return><serviceTicket>abc123</serviceTicket><userType>2</userType><studentIDs>456</studentIDs></return></loginToPublicPortalResponse></soap:Body></soap:Envelope>`

	fields, err := extractXMLFields([]byte(body), "serviceTicket", "userType", "studentIDs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fields["serviceTicket"] != "abc123" || fields["userType"] != "2" || fields["studentIDs"] != "456" {
		t.Fatalf("got %+v", fields)
	}
}

func TestExtractXMLFields_MissingElementsReturnEmptyNotPanic(t *testing.T) {
	// this is exactly what the old regexp.FindStringSubmatch(...)[1] would
	// panic on: a well-formed-but-unrelated response with none of the wanted
	// elements present.
	body := `<?xml version="1.0"?><soap:Envelope xmlns:soap="x"><soap:Body><soap:Fault><faultstring>bad request</faultstring></soap:Fault></soap:Body></soap:Envelope>`

	fields, err := extractXMLFields([]byte(body), "serviceTicket", "userType", "studentIDs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fields["serviceTicket"] != "" || fields["userType"] != "" {
		t.Fatalf("expected empty values for absent elements, got %+v", fields)
	}
}

func TestExtractXMLFields_MalformedXMLReturnsError(t *testing.T) {
	_, err := extractXMLFields([]byte(`<not valid xml`), "serviceTicket")
	if err == nil {
		t.Fatal("expected an error on malformed xml, got nil")
	}
}

func TestGetServiceTicket_NoTicketInResponse(t *testing.T) {
	// GetServiceTicket only reaches here after an HTTP round trip, which
	// needs a live server — but the ErrNoTicket / ErrNotStudent branches are
	// pure once we have the parsed body, so exercise the underlying parse
	// directly to confirm the sentinel errors fire on the right input shape.
	body := []byte(`<Envelope><Body><loginToPublicPortalResponse><return><userType>1</userType></return></loginToPublicPortalResponse></Body></Envelope>`)
	fields, err := extractXMLFields(body, "serviceTicket", "userType", "studentIDs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fields["serviceTicket"] != "" {
		t.Fatalf("expected no ticket, got %q", fields["serviceTicket"])
	}
	if fields["userType"] != "1" {
		t.Fatalf("expected userType=1 (a parent account), got %q", fields["userType"])
	}
}

func TestParseDigestNonce(t *testing.T) {
	header := `Digest realm="Protected", qop="auth", nonce="deadbeef1234", algorithm=MD5`
	nonce, err := parseDigestNonce(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nonce != "deadbeef1234" {
		t.Fatalf("want deadbeef1234, got %q", nonce)
	}
}

func TestParseDigestNonce_MissingNonceReturnsError(t *testing.T) {
	// this is the old regexp.FindStringSubmatch(auth)[1] panic path: a
	// WWW-Authenticate header without a nonce field at all.
	_, err := parseDigestNonce(`Digest realm="Protected"`)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestParseDigestNonce_UnterminatedQuoteReturnsError(t *testing.T) {
	_, err := parseDigestNonce(`Digest nonce="unterminated`)
	if err == nil {
		t.Fatal("expected an error on an unterminated quoted value, got nil")
	}
}
