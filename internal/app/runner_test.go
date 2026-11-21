package app

import (
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/osmargm1202/orgmdns/internal/cloudflare"
	"github.com/osmargm1202/orgmdns/internal/config"
	"github.com/osmargm1202/orgmdns/internal/logger"
	"github.com/osmargm1202/orgmdns/internal/notify"
)

type fakeDNS struct {
	ips      map[string]string
	failures map[string]bool
	updates  int
}

func (f *fakeDNS) GetDNSRecordByName(name string) (*cloudflare.DNSRecord, error) {
	if name == "missing" {
		return nil, errors.New("not found")
	}
	return &cloudflare.DNSRecord{ID: name, Content: f.ips[name]}, nil
}
func (f *fakeDNS) UpdateDNSRecordIP(id, ip string) error {
	if f.failures[id] {
		return errors.New("API failure")
	}
	f.ips[id] = ip
	f.updates++
	return nil
}

type fakeNotifier struct {
	attempts int
	fail     bool
	messages [][]notify.Event
}

func (f *fakeNotifier) SendSummary(events []notify.Event) error {
	f.attempts++
	if f.fail {
		return errors.New("SMTP down")
	}
	f.messages = append(f.messages, append([]notify.Event(nil), events...))
	return nil
}
func testRunner() (*Runner, *fakeDNS, *fakeNotifier) {
	dns := &fakeDNS{ips: map[string]string{"a": "old", "b": "old"}, failures: map[string]bool{}}
	mail := &fakeNotifier{}
	r := &Runner{config: &config.Config{RecordNames: []string{"a", "b"}}, logger: &logger.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}, cf: dns, notifier: mail, checkInternet: func() bool { return true }, publicIP: func() (string, error) { return "new", nil }}
	return r, dns, mail
}
func TestOneSummaryAndQuietCycle(t *testing.T) {
	r, dns, mail := testRunner()
	down := time.Now().Add(-time.Minute)
	r.internetDown = true
	r.disconnectedAt = &down
	r.runCycle()
	if mail.attempts != 1 || len(mail.messages[0]) != 4 || dns.updates != 2 {
		t.Fatalf("summary=%+v updates=%d", mail, dns.updates)
	}
	r.runCycle()
	if mail.attempts != 1 {
		t.Fatal("unchanged cycle sent mail")
	}
}
func TestPartialFailureContinuesAndRetries(t *testing.T) {
	r, dns, mail := testRunner()
	dns.failures["a"] = true
	r.runCycle()
	if dns.updates != 1 || mail.attempts != 1 {
		t.Fatal("did not continue after failure")
	}
	if mail.messages[0][1].Kind != "Error DNS" || mail.messages[0][2].Kind != "DNS actualizado" {
		t.Fatal(mail.messages)
	}
	dns.failures["a"] = false
	r.runCycle()
	if dns.updates != 2 || len(mail.messages[1]) != 1 {
		t.Fatal("failed DNS update not retried")
	}
}
func TestSMTPFailureRetainsConfirmedUpdates(t *testing.T) {
	r, dns, mail := testRunner()
	mail.fail = true
	r.runCycle()
	r.runCycle()
	if len(r.pending) != 3 || dns.updates != 2 {
		t.Fatal("lost or duplicated events")
	}
	mail.fail = false
	r.runCycle()
	if len(mail.messages) != 1 || len(mail.messages[0]) != 3 || len(r.pending) != 0 {
		t.Fatal("summary not recovered")
	}
}
func TestOfflineAndIPFailure(t *testing.T) {
	r, dns, mail := testRunner()
	r.checkInternet = func() bool { return false }
	r.runCycle()
	if mail.attempts != 0 || dns.updates != 0 {
		t.Fatal("offline side effect")
	}
	r.checkInternet = func() bool { return true }
	r.publicIP = func() (string, error) { return "", errors.New("down") }
	r.runCycle()
	if mail.attempts != 1 || len(mail.messages[0]) != 2 || dns.updates != 0 {
		t.Fatal("reconnection and IP error not grouped")
	}
}

func TestSharedQueueGroupsServerDNSAndSurvivesRestart(t *testing.T) {
	r, dns, mail := testRunner()
	r.config.NotificationDir = t.TempDir()
	event := notify.Event{Time: time.Now(), Kind: "Servidor iniciado", Detail: "Evento de orgmserver"}
	if err := notify.Enqueue(r.config.NotificationDir, event); err != nil {
		t.Fatal(err)
	}
	r.config.NotificationSettle = time.Minute
	r.runCycle()
	if mail.attempts != 0 || dns.updates != 2 {
		t.Fatal("quiet window ignored")
	}
	r.config.NotificationSettle = 0
	mail.fail = true
	r.flush()
	// Simulate a restart: all events, including confirmed DNS updates, are on disk.
	r.pending = nil
	r.startupQueued = false
	mail.fail = false
	r.runCycle()
	if len(mail.messages) != 1 || len(mail.messages[0]) != 4 || dns.updates != 2 {
		t.Fatal("events lost/duplicated", mail.messages)
	}
	queued, err := notify.ReadQueue(r.config.NotificationDir)
	if err != nil || len(queued) != 0 {
		t.Fatal("delivered files not acknowledged", err)
	}
}
