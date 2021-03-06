package mexinfo

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	customsearch "google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

const (
	version                     = "v0"
	slackRequestTimestampHeader = "X-Slack-Request-Timestamp"
	slackSignatureHeader        = "X-Slack-Signature"
)

type oldTimeStampError struct {
	s string
}

func (e *oldTimeStampError) Error() string {
	return e.s
}

// Message is slack message
type Message struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
	UnfurlLinks  bool   `json:"unfurl_links"`
}

// Result is customsearch result
type Result struct {
	Position int64
	Result   *customsearch.Result
}

// MexSearch brings Mexican information to the Sumo world
func MexSearch(w http.ResponseWriter, r *http.Request) {
	setup(r.Context())

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("Couldn't read request body: %v", err)
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	if r.Method != "POST" {
		http.Error(w, "Only POST requests are accepted", 405)
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Couldn't parse form", 400)
		log.Fatalf("ParseForm: %v", err)
	}

	// Reset r.Body as ParseForm depletes it by reading the io.ReadCloser.
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	result, err := verifyWebHook(r)
	if err != nil {
		log.Fatalf("verifyWebhook: %v", err)
	}
	if !result {
		log.Fatalf("signatures did not match.")
	}

	if len(r.Form["text"]) == 0 {
		log.Fatalf("emtpy text in form")
	}

	// todo: search
	query := strings.Join(r.Form["text"], " ")
	log.Printf("query %s", query)

	msg, err := makeSearchRequest(r.Context(), query)
	if err != nil {
		log.Fatalf("makeSearchRequest: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		log.Fatalf("json.Marshal: %v", err)
	}
	log.Printf("send message %v", msg)
}

func verifyWebHook(r *http.Request) (bool, error) {
	slackSigningSecret := os.Getenv("SLACK_SECRET")
	timeStamp := r.Header.Get(slackRequestTimestampHeader)
	slackSignature := r.Header.Get(slackSignatureHeader)

	t, err := strconv.ParseInt(timeStamp, 10, 64)
	if err != nil {
		return false, fmt.Errorf("strconv.ParseInt(%s): %v", timeStamp, err)
	}

	if ageOk, age := checkTimestamp(t); !ageOk {
		return false, &oldTimeStampError{fmt.Sprintf("checkTimestamp(%v): %v %v", t, ageOk, age)}
	}

	if timeStamp == "" || slackSignature == "" {
		return false, fmt.Errorf("either timeStamp or signature headers were blank")
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return false, fmt.Errorf("ioutil.ReadAll(%v): %v", r.Body, err)
	}

	// Reset the body so other calls won't fail.
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	baseString := fmt.Sprintf("%s:%s:%s", version, timeStamp, body)

	signature := getSignature([]byte(baseString), []byte(slackSigningSecret))

	trimmed := strings.TrimPrefix(slackSignature, fmt.Sprintf("%s=", version))
	signatureInHeader, err := hex.DecodeString(trimmed)

	if err != nil {
		return false, fmt.Errorf("hex.DecodeString(%v): %v", trimmed, err)
	}

	return hmac.Equal(signature, signatureInHeader), nil
}

func makeSearchRequest(ctx context.Context, query string) (*Message, error) {
	apiKey := os.Getenv("SEARCH_API_KEY")
	id := os.Getenv("SEARCH_ID")
	q := "メキシコ " + query

	customsearchService, err := customsearch.NewService(ctx, option.WithAPIKey(apiKey))
	search := customsearchService.Cse.List(q)
	search.Cx(id)

	search.Start(1)
	call, err := search.Do()
	if err != nil {
		return nil, err
	}
	if len(call.Items) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return formatSlackMessage(call.Items[0].Link)
}

func getSignature(base []byte, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(base)

	return h.Sum(nil)
}

func checkTimestamp(timeStamp int64) (bool, time.Duration) {
	t := time.Since(time.Unix(timeStamp, 0))

	return t.Minutes() <= 5, t
}
