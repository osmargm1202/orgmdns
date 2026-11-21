package notify

import (
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"testing"
	"time"
)

func TestMIMESummaryAndEscaping(t *testing.T) {
	msg, err := BuildMessage("from@example.test", "to@example.test", []Event{{Time: time.Now(), Kind: "DNS actualizado", Record: "<script>alert(1)</script>", OldIP: "192.0.2.1", NewIP: "192.0.2.2", Detail: "Confirmación"}})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := mail.ReadMessage(strings.NewReader(string(msg)))
	if err != nil {
		t.Fatal(err)
	}
	kind, params, err := mime.ParseMediaType(parsed.Header.Get("Content-Type"))
	if err != nil || kind != "multipart/alternative" {
		t.Fatal(kind, err)
	}
	reader := multipart.NewReader(parsed.Body, params["boundary"])
	for _, expected := range []string{"text/plain", "text/html"} {
		part, err := reader.NextPart()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(part.Header.Get("Content-Type"), expected) {
			t.Fatal(part.Header)
		}
		body, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "192.0.2.2") || !strings.Contains(string(body), "Confirmación") {
			t.Fatal(string(body))
		}
		if expected == "text/html" && (strings.Contains(string(body), "<script>") || !strings.Contains(string(body), "&lt;script&gt;")) {
			t.Fatal("HTML not escaped")
		}
	}
	if _, err := reader.NextPart(); err != io.EOF {
		t.Fatal("unexpected part", err)
	}
}
