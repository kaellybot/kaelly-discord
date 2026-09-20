package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

// DiscordRequest is one REST call the session attempted.
type DiscordRequest struct {
	Method string
	Path   string // URL path, e.g. /api/v9/webhooks/<app>/<token>/messages/@original
	Body   []byte // JSON payload; for multipart requests, the payload_json part
	Files  []string
}

// Decode unmarshals the JSON payload into v.
func (r DiscordRequest) Decode(t *testing.T, v any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(r.Body, v), "body: %s", r.Body)
}

// DiscordResponse is a canned answer for requests matching Method and a Path
// suffix/substring.
type DiscordResponse struct {
	Method       string // empty matches any method
	PathContains string
	Status       int
	Body         string
}

// FakeDiscord records every REST call made through the session returned by
// NewDiscordSession. Unmatched requests get 200 with "{}".
type FakeDiscord struct {
	mu        sync.Mutex
	requests  []DiscordRequest
	responses []DiscordResponse
}

// NewDiscordSession returns a session whose HTTP client never leaves the
// process, plus the recorder observing it.
func NewDiscordSession(t *testing.T) (*discordgo.Session, *FakeDiscord) {
	t.Helper()
	session, err := discordgo.New("Bot test-token")
	require.NoError(t, err)
	fake := &FakeDiscord{}
	session.Client = &http.Client{Transport: fake}
	session.MaxRestRetries = 0
	session.ShouldRetryOnRateLimit = false
	session.State.User = &discordgo.User{ID: "bot-id", Username: "KaellyTest"}
	return session, fake
}

// Respond registers a canned response; the latest matching registration wins.
func (f *FakeDiscord) Respond(response DiscordResponse) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.responses = append(f.responses, response)
}

// Requests returns a snapshot of every recorded request, in order.
func (f *FakeDiscord) Requests() []DiscordRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]DiscordRequest(nil), f.requests...)
}

// Last returns the most recent request, failing the test if there is none.
func (f *FakeDiscord) Last(t *testing.T) DiscordRequest {
	t.Helper()
	requests := f.Requests()
	require.NotEmpty(t, requests, "no Discord request was made")
	return requests[len(requests)-1]
}

// LastEdit decodes the most recent request body as a webhook edit payload,
// the shape of InteractionResponseEdit.
func (f *FakeDiscord) LastEdit(t *testing.T) WebhookEditPayload {
	t.Helper()
	var payload WebhookEditPayload
	f.Last(t).Decode(t, &payload)
	return payload
}

// WebhookEditPayload mirrors the JSON sent by InteractionResponseEdit.
type WebhookEditPayload struct {
	Content    *string                   `json:"content"`
	Embeds     []*discordgo.MessageEmbed `json:"embeds"`
	Components []json.RawMessage         `json:"components"`
}

func (f *FakeDiscord) RoundTrip(req *http.Request) (*http.Response, error) {
	recorded := DiscordRequest{Method: req.Method, Path: req.URL.Path}
	if req.Body != nil {
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		recorded.Body, recorded.Files = splitMultipart(req.Header.Get("Content-Type"), raw)
	}

	f.mu.Lock()
	f.requests = append(f.requests, recorded)
	status, body := http.StatusOK, "{}"
	for _, r := range slices.Backward(f.responses) {
		if (r.Method == "" || r.Method == req.Method) && strings.Contains(req.URL.Path, r.PathContains) {
			status, body = r.Status, r.Body
			break
		}
	}
	f.mu.Unlock()

	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func splitMultipart(contentType string, raw []byte) ([]byte, []string) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		return raw, nil
	}

	var payload []byte
	var files []string
	reader := multipart.NewReader(bytes.NewReader(raw), params["boundary"])
	for {
		part, errPart := reader.NextPart()
		if errPart != nil {
			break
		}
		content, _ := io.ReadAll(part)
		if part.FormName() == "payload_json" {
			payload = content
		} else {
			files = append(files, part.FileName())
		}
	}
	return payload, files
}
