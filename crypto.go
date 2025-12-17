package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/unmistakenly/PSGradeUtility/powerschool"
	"github.com/unmistakenly/PSGradeUtility/powerschool/crypto"
)

var httpClient = &http.Client{Timeout: 5 * time.Second}

var (
	ErrNoTicket   = errors.New("serviceTicket not found in response body\nare you sure your password is correct?")
	ErrNotStudent = errors.New("parent accounts are unsupported, please sign in using your own account")
)

func GetServiceTicket(username, password string) (string, string, error) {
	nonce, nonceDate := crypto.Nonce(8, true)
	body := strings.NewReplacer(
		"{nonce}", nonce,
		"{nonceDate}", nonceDate,
		"{username}", username,
		"{password}", password,
	).Replace(powerschool.PortalServiceLoginTemplate)

	authReq, _ := http.NewRequest(http.MethodPost, "https://myps.horrycountyschools.net/pearson-rest/services/PublicPortalService", strings.NewReader(body))

	authResp, err := httpClient.Do(authReq)
	if err != nil {
		return "", "", fmt.Errorf("making request: %w", err)
	}
	defer authResp.Body.Close()

	respBody := bytes.NewBuffer(make([]byte, 0, 1800))
	if _, err = io.Copy(respBody, io.LimitReader(authResp.Body, 1800)); err != nil {
		return "", "", fmt.Errorf("reading response body: %w", err)
	}
	sRespBody := respBody.String()

	ticketMatches := regexp.MustCompile(`\<serviceTicket\>(.+)\<\/serviceTicket\>`).FindStringSubmatch(sRespBody)
	if len(ticketMatches) < 2 {
		return "", "", ErrNoTicket
	} else if regexp.MustCompile(`\<userType\>(.+)\<\/userType\>`).FindStringSubmatch(sRespBody)[1] != "2" {
		return "", "", ErrNotStudent
	}

	return ticketMatches[1],
		regexp.MustCompile(`\<studentIDs\>(.+)\<\/studentIDs\>`).FindStringSubmatch(sRespBody)[1],
		nil
}

// the caller is responsible for closing the response body
func GetFullData(ticket, studentID string) (io.ReadCloser, error) {
	const portalURL = "https://myps.horrycountyschools.net/pearson-rest/services/PublicPortalServiceJSON?response=application/json"

	// first, you have to make the initial request and get a 401
	body := strings.NewReplacer(
		"{ticket}", ticket,
		"{studentID}", studentID,
	).Replace(powerschool.DataRequestTemplate)

	authReq, _ := http.NewRequest(http.MethodPost, portalURL, strings.NewReader(body))
	authResp, err := httpClient.Do(authReq)
	if err != nil {
		return nil, fmt.Errorf("making initial unauthorized request: %w", err)
	}
	authResp.Body.Close()

	// then, get the nonce from WWW-Authenticate and generate cnonce + response
	auth := authResp.Header.Get("WWW-Authenticate")
	nonce := regexp.MustCompile(`nonce="(.+)"`).FindStringSubmatch(auth)[1]

	cnonce, _ := crypto.Nonce(16, false)
	cresponse := crypto.DigestResponse(nonce, cnonce)

	req, _ := http.NewRequest(http.MethodPost, portalURL, strings.NewReader(body))
	req.Header.Set(
		"Authorization",
		fmt.Sprintf(
			`Digest username="pearson", realm="Protected", nonce="%s", uri="/pearson-rest/services/PublicPortalServiceJSON?response=application/json", response="%s", cnonce="%s", nc=00000001, qop="auth"`,
			nonce, cresponse, cnonce,
		),
	)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making authorized request: %w", err)
	}

	return resp.Body, nil
}
