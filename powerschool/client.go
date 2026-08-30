package powerschool

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/unmistakenly/PSGradeUtility/powerschool/crypto"
)

var (
	ErrNoTicket   = errors.New("serviceTicket not found in response body\nare you sure your password is correct?")
	ErrNotStudent = errors.New("parent accounts are unsupported, please sign in using your own account")
	ErrNoNonce    = errors.New("no nonce found in WWW-Authenticate header")
)

// Client talks to a single PowerSchool public-portal instance.
type Client struct {
	BaseURL    string // e.g. https://myps.<district>.org, no trailing slash
	HTTPClient *http.Client
}

// NewClient returns a Client with a sane default timeout. baseURL is required —
// there is no default district baked in (rule-10: never ship a real one).
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// extractXMLFields walks the raw XML by token (not a fixed struct) and returns
// the text content of the first element matching each requested local name,
// wherever it appears in the tree. The old code found <serviceTicket> etc. by
// regex scanning the raw body regardless of nesting/namespace — we don't have
// a captured sample of the real envelope's exact nesting to hand-model a
// struct against, so token-scanning by local name is the encoding/xml
// equivalent of that same "find it anywhere" behavior, without the panic risk
// of an unchecked regex[1] index.
func extractXMLFields(data []byte, names ...string) (map[string]string, error) {
	want := make(map[string]bool, len(names))
	for _, n := range names {
		want[n] = true
	}
	found := make(map[string]string, len(names))

	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parsing xml: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok || !want[start.Name.Local] {
			continue
		}
		var text string
		if err := dec.DecodeElement(&text, &start); err != nil {
			return nil, fmt.Errorf("parsing xml element %s: %w", start.Name.Local, err)
		}
		if _, already := found[start.Name.Local]; !already {
			found[start.Name.Local] = text
		}
	}
	return found, nil
}

func (c *Client) GetServiceTicket(username, password string) (ticket, studentID string, err error) {
	nonce, nonceDate := crypto.Nonce(8, true)
	body := strings.NewReplacer(
		"{nonce}", nonce,
		"{nonceDate}", nonceDate,
		"{username}", username,
		"{password}", password,
	).Replace(PortalServiceLoginTemplate)

	req, err := http.NewRequest(
		http.MethodPost,
		c.BaseURL+"/pearson-rest/services/PublicPortalService",
		strings.NewReader(body),
	)
	if err != nil {
		return "", "", fmt.Errorf("building request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody := bytes.NewBuffer(make([]byte, 0, 1800))
	if _, err = io.Copy(respBody, io.LimitReader(resp.Body, 1800)); err != nil {
		return "", "", fmt.Errorf("reading response body: %w", err)
	}

	fields, err := extractXMLFields(respBody.Bytes(), "serviceTicket", "userType", "studentIDs")
	if err != nil {
		return "", "", err
	}

	ticket = fields["serviceTicket"]
	if ticket == "" {
		return "", "", ErrNoTicket
	}
	if fields["userType"] != "2" {
		return "", "", ErrNotStudent
	}

	return ticket, fields["studentIDs"], nil
}

// the caller is responsible for closing the response body
func (c *Client) GetFullData(ticket, studentID string) (io.ReadCloser, error) {
	portalURL := c.BaseURL + "/pearson-rest/services/PublicPortalServiceJSON?response=application/json"

	// step 1: an intentionally-unauthenticated request, to harvest a nonce off
	// the 401's WWW-Authenticate header.
	body := strings.NewReplacer(
		"{ticket}", ticket,
		"{studentID}", studentID,
	).Replace(DataRequestTemplate)

	authReq, err := http.NewRequest(http.MethodPost, portalURL, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building initial request: %w", err)
	}
	authResp, err := c.HTTPClient.Do(authReq)
	if err != nil {
		return nil, fmt.Errorf("making initial unauthorized request: %w", err)
	}
	authResp.Body.Close()

	nonce, err := parseDigestNonce(authResp.Header.Get("WWW-Authenticate"))
	if err != nil {
		return nil, err
	}

	cnonce, _ := crypto.Nonce(16, false)
	cresponse := crypto.DigestResponse(nonce, cnonce)

	// step 2: the real, Digest-authenticated request.
	req, err := http.NewRequest(http.MethodPost, portalURL, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building authorized request: %w", err)
	}
	req.Header.Set(
		"Authorization",
		fmt.Sprintf(
			`Digest username="pearson", realm="Protected", nonce="%s", uri="/pearson-rest/services/PublicPortalServiceJSON?response=application/json", response="%s", cnonce="%s", nc=00000001, qop="auth"`,
			nonce, cresponse, cnonce,
		),
	)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making authorized request: %w", err)
	}

	return resp.Body, nil
}

// parseDigestNonce pulls nonce="..." out of a WWW-Authenticate Digest header.
// This is an HTTP header value, not XML — parsed as one, bounds-checked, no
// regex-into-[1] panic risk.
func parseDigestNonce(header string) (string, error) {
	const key = `nonce="`
	i := strings.Index(header, key)
	if i < 0 {
		return "", ErrNoNonce
	}
	rest := header[i+len(key):]
	j := strings.IndexByte(rest, '"')
	if j < 0 {
		return "", ErrNoNonce
	}
	return rest[:j], nil
}
