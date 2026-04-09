package mail

import (
	"context"
	"errors"
	"math/rand"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "no accounts",
			cfg:  Config{},
		},
		{
			name: "missing host",
			cfg: Config{
				Accounts: []AccountConfig{{Port: 587}},
			},
		},
		{
			name: "missing port",
			cfg: Config{
				Accounts: []AccountConfig{{Host: "smtp.example.com"}},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := New(test.cfg); err == nil {
				t.Fatalf("expected config validation error")
			}
		})
	}
}

func TestSendValidation(t *testing.T) {
	t.Parallel()

	service := newTestService(t, Config{
		Accounts: []AccountConfig{
			{
				Name:       "primary",
				Host:       "smtp-1.example.com",
				Port:       587,
				Username:   "user1@example.com",
				Password:   "secret",
				Encryption: EncryptionSTARTTLS,
				From:       Address{Name: "Primary", Email: "noreply1@example.com"},
			},
		},
	})

	tests := []struct {
		name string
		msg  Message
	}{
		{
			name: "missing recipients",
			msg: Message{
				Subject:  "hello",
				TextBody: "world",
			},
		},
		{
			name: "missing body",
			msg: Message{
				To:      []string{"user@example.com"},
				Subject: "hello",
			},
		},
		{
			name: "invalid attachment",
			msg: Message{
				To:       []string{"user@example.com"},
				Subject:  "hello",
				TextBody: "world",
				Attachments: []Attachment{
					{Filename: "broken.txt"},
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := service.Send(context.Background(), test.msg); err == nil {
				t.Fatalf("expected validation error")
			}
		})
	}
}

func TestRoundRobinRotation(t *testing.T) {
	t.Parallel()

	mock := &mockSender{}
	service := newTestService(t, Config{
		Accounts: []AccountConfig{
			testAccount("primary", "smtp-1.example.com", "noreply1@example.com"),
			testAccount("backup", "smtp-2.example.com", "noreply2@example.com"),
		},
		EnableRotation:   true,
		RotationStrategy: RotationStrategyRoundRobin,
	}, mock)

	msg := Message{
		To:       []string{"user@example.com"},
		Subject:  "rotation",
		TextBody: "hello",
	}
	for range 3 {
		if err := service.Send(context.Background(), msg); err != nil {
			t.Fatalf("send failed: %v", err)
		}
	}

	got := mock.accountNames()
	want := []string{"primary", "backup", "primary"}
	assertSliceEqual(t, got, want)
}

func TestRandomRotationUsesMultipleAccounts(t *testing.T) {
	t.Parallel()

	mock := &mockSender{}
	service := newTestService(t, Config{
		Accounts: []AccountConfig{
			testAccount("primary", "smtp-1.example.com", "noreply1@example.com"),
			testAccount("backup", "smtp-2.example.com", "noreply2@example.com"),
		},
		EnableRotation:   true,
		RotationStrategy: RotationStrategyRandom,
	}, mock)
	service.random = rand.New(rand.NewSource(1))

	msg := Message{
		To:       []string{"user@example.com"},
		Subject:  "rotation",
		TextBody: "hello",
	}
	for range 8 {
		if err := service.Send(context.Background(), msg); err != nil {
			t.Fatalf("send failed: %v", err)
		}
	}

	seen := map[string]struct{}{}
	for _, name := range mock.accountNames() {
		seen[name] = struct{}{}
	}
	if len(seen) < 2 {
		t.Fatalf("expected random rotation to use multiple accounts, got %v", mock.accountNames())
	}
}

func TestNoneStrategyUsesFirstAccount(t *testing.T) {
	t.Parallel()

	mock := &mockSender{}
	service := newTestService(t, Config{
		Accounts: []AccountConfig{
			testAccount("primary", "smtp-1.example.com", "noreply1@example.com"),
			testAccount("backup", "smtp-2.example.com", "noreply2@example.com"),
		},
		EnableRotation: false,
	}, mock)

	msg := Message{
		To:       []string{"user@example.com"},
		Subject:  "rotation",
		TextBody: "hello",
	}
	for range 3 {
		if err := service.Send(context.Background(), msg); err != nil {
			t.Fatalf("send failed: %v", err)
		}
	}

	want := []string{"primary", "primary", "primary"}
	assertSliceEqual(t, mock.accountNames(), want)
}

func TestRetryableFailover(t *testing.T) {
	t.Parallel()

	mock := &mockSender{
		fn: func(call int, account AccountConfig, from string, recipients []string, data []byte) error {
			if call == 0 {
				return newRetryableError("smtp dial failed", errors.New("connection reset"))
			}
			return nil
		},
	}
	service := newTestService(t, Config{
		Accounts: []AccountConfig{
			testAccount("primary", "smtp-1.example.com", "noreply1@example.com"),
			testAccount("backup", "smtp-2.example.com", "noreply2@example.com"),
		},
		EnableRotation: false,
	}, mock)

	err := service.Send(context.Background(), Message{
		To:       []string{"user@example.com"},
		Subject:  "retry",
		TextBody: "hello",
	})
	if err != nil {
		t.Fatalf("expected failover to succeed, got %v", err)
	}

	assertSliceEqual(t, mock.accountNames(), []string{"primary", "backup"})
}

func TestPermanentFailureDoesNotFailover(t *testing.T) {
	t.Parallel()

	mock := &mockSender{
		fn: func(call int, account AccountConfig, from string, recipients []string, data []byte) error {
			return newPermanentError("smtp rcpt failed", &textproto.Error{Code: 550, Msg: "user unknown"})
		},
	}
	service := newTestService(t, Config{
		Accounts: []AccountConfig{
			testAccount("primary", "smtp-1.example.com", "noreply1@example.com"),
			testAccount("backup", "smtp-2.example.com", "noreply2@example.com"),
		},
		EnableRotation: false,
	}, mock)

	err := service.Send(context.Background(), Message{
		To:       []string{"user@example.com"},
		Subject:  "retry",
		TextBody: "hello",
	})
	if err == nil {
		t.Fatalf("expected permanent error")
	}
	assertSliceEqual(t, mock.accountNames(), []string{"primary"})
}

func TestCustomHTMLAndAttachments(t *testing.T) {
	t.Parallel()

	filePath := filepath.Join(t.TempDir(), "report.txt")
	if err := os.WriteFile(filePath, []byte("path-data"), 0o644); err != nil {
		t.Fatalf("write temp attachment failed: %v", err)
	}

	mock := &mockSender{}
	service := newTestService(t, Config{
		Accounts: []AccountConfig{
			testAccount("primary", "smtp-1.example.com", "noreply1@example.com"),
		},
		DefaultFrom: Address{Name: "Template Team", Email: "team@example.com"},
	}, mock)

	err := service.Send(context.Background(), Message{
		To:       []string{"to@example.com"},
		Cc:       []string{"cc@example.com"},
		Bcc:      []string{"bcc@example.com"},
		ReplyTo:  []string{"reply@example.com"},
		Subject:  "HTML",
		TextBody: "fallback text",
		HTMLBody: "<h1>Hello</h1>",
		Headers: map[string]string{
			"X-Trace-ID": "trace-001",
		},
		Attachments: []Attachment{
			{Filename: "inline.txt", Data: []byte("hello")},
			{Path: filePath},
		},
	})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}

	if len(mock.calls) != 1 {
		t.Fatalf("expected one send call, got %d", len(mock.calls))
	}
	call := mock.calls[0]
	if call.from != "team@example.com" {
		t.Fatalf("unexpected envelope from: %s", call.from)
	}
	assertSliceEqual(t, call.recipients, []string{"to@example.com", "cc@example.com", "bcc@example.com"})

	if !strings.Contains(call.data, "Cc: <cc@example.com>") {
		t.Fatalf("expected Cc header in mime message: %s", call.data)
	}
	if strings.Contains(call.data, "Bcc:") {
		t.Fatalf("expected Bcc header to stay out of mime message: %s", call.data)
	}
	if !strings.Contains(call.data, "Reply-To: <reply@example.com>") {
		t.Fatalf("expected Reply-To header in mime message")
	}
	if !strings.Contains(call.data, "X-Trace-Id: trace-001") {
		t.Fatalf("expected custom header in mime message")
	}
	if !strings.Contains(call.data, "Content-Type: multipart/mixed;") {
		t.Fatalf("expected multipart/mixed body")
	}
	if !strings.Contains(call.data, "Content-Type: multipart/alternative;") {
		t.Fatalf("expected multipart/alternative subpart")
	}
	if !strings.Contains(call.data, `Content-Disposition: attachment; filename="inline.txt"`) {
		t.Fatalf("expected inline attachment header")
	}
	if !strings.Contains(call.data, `Content-Disposition: attachment; filename="report.txt"`) {
		t.Fatalf("expected path attachment header")
	}
	if !strings.Contains(call.data, "aGVsbG8=") {
		t.Fatalf("expected inline attachment body to be base64 encoded")
	}
	if !strings.Contains(call.data, "cGF0aC1kYXRh") {
		t.Fatalf("expected file attachment body to be base64 encoded")
	}
}

func TestSendTemplateRendersBuiltins(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template TemplateName
		data     map[string]any
		want     []string
	}{
		{
			name:     "welcome",
			template: TemplateWelcome,
			data: map[string]any{
				"recipient_name": "Alice",
				"action_url":     "https://example.com/open",
			},
			want: []string{"Welcome aboard", "Alice", "https://example.com/open"},
		},
		{
			name:     "verify code",
			template: TemplateVerifyCode,
			data: map[string]any{
				"code":       "123456",
				"expires_in": "10 minutes",
			},
			want: []string{"Verification code", "123456", "10 minutes"},
		},
		{
			name:     "reset password",
			template: TemplateResetPassword,
			data: map[string]any{
				"reset_url": "https://example.com/reset",
			},
			want: []string{"Reset your password", "https://example.com/reset"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockSender{}
			service := newTestService(t, Config{
				Accounts: []AccountConfig{
					testAccount("primary", "smtp-1.example.com", "noreply1@example.com"),
				},
			}, mock)

			err := service.SendTemplate(context.Background(), TemplateMessage{
				Message: Message{
					To:      []string{"user@example.com"},
					Subject: "Template",
				},
				Template: test.template,
				Data:     test.data,
			})
			if err != nil {
				t.Fatalf("send template failed: %v", err)
			}
			if len(mock.calls) != 1 {
				t.Fatalf("expected one send call")
			}
			for _, want := range test.want {
				if !strings.Contains(mock.calls[0].data, want) {
					t.Fatalf("expected mime message to contain %q, got %s", want, mock.calls[0].data)
				}
			}
		})
	}
}

func TestSendTemplateRejectsHTMLMix(t *testing.T) {
	t.Parallel()

	service := newTestService(t, Config{
		Accounts: []AccountConfig{
			testAccount("primary", "smtp-1.example.com", "noreply1@example.com"),
		},
	})

	err := service.SendTemplate(context.Background(), TemplateMessage{
		Message: Message{
			To:       []string{"user@example.com"},
			Subject:  "Template",
			HTMLBody: "<p>custom</p>",
		},
		Template: TemplateWelcome,
	})
	if err == nil {
		t.Fatalf("expected template/html conflict error")
	}
}

func newTestService(t *testing.T, cfg Config, senders ...smtpSender) *Service {
	t.Helper()

	service, err := New(cfg)
	if err != nil {
		t.Fatalf("new service failed: %v", err)
	}
	if len(senders) > 0 && senders[0] != nil {
		service.sender = senders[0]
	}
	return service
}

func testAccount(name, host, from string) AccountConfig {
	return AccountConfig{
		Name:       name,
		Host:       host,
		Port:       587,
		Username:   "auth@" + host,
		Password:   "secret",
		Encryption: EncryptionSTARTTLS,
		From:       Address{Name: strings.Title(name), Email: from},
	}
}

func assertSliceEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("unexpected slice length: got=%v want=%v", got, want)
	}
	for index := range got {
		if got[index] != want[index] {
			t.Fatalf("unexpected slice content: got=%v want=%v", got, want)
		}
	}
}

type mockSender struct {
	mu    sync.Mutex
	calls []mockCall
	fn    func(call int, account AccountConfig, from string, recipients []string, data []byte) error
}

type mockCall struct {
	account    string
	from       string
	recipients []string
	data       string
}

func (m *mockSender) Send(ctx context.Context, account AccountConfig, envelopeFrom string, recipients []string, data []byte, dialTimeout, sendTimeout time.Duration, localName string) error {
	_ = ctx
	_ = dialTimeout
	_ = sendTimeout
	_ = localName

	m.mu.Lock()
	callIndex := len(m.calls)
	m.calls = append(m.calls, mockCall{
		account:    account.displayName(),
		from:       envelopeFrom,
		recipients: append([]string(nil), recipients...),
		data:       string(append([]byte(nil), data...)),
	})
	fn := m.fn
	m.mu.Unlock()

	if fn != nil {
		return fn(callIndex, account, envelopeFrom, recipients, data)
	}
	return nil
}

func (m *mockSender) accountNames() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]string, 0, len(m.calls))
	for _, call := range m.calls {
		result = append(result, call.account)
	}
	return result
}
