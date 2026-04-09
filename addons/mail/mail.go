package mail

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	stdmail "net/mail"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultDialTimeout = 10 * time.Second
	defaultSendTimeout = 15 * time.Second
)

type RotationStrategy string

const (
	RotationStrategyNone       RotationStrategy = "none"
	RotationStrategyRoundRobin RotationStrategy = "round_robin"
	RotationStrategyRandom     RotationStrategy = "random"
)

type EncryptionMode string

const (
	EncryptionNone     EncryptionMode = "none"
	EncryptionSTARTTLS EncryptionMode = "starttls"
	EncryptionSSLTLS   EncryptionMode = "ssl_tls"
)

type TemplateName string

const (
	TemplateWelcome       TemplateName = "welcome"
	TemplateVerifyCode    TemplateName = "verify_code"
	TemplateResetPassword TemplateName = "reset_password"
)

type Address struct {
	Name  string
	Email string
}

type Config struct {
	Accounts         []AccountConfig
	RotationStrategy RotationStrategy
	EnableRotation   bool
	DialTimeout      time.Duration
	SendTimeout      time.Duration
	LocalName        string
	DefaultFrom      Address
}

type AccountConfig struct {
	Name               string
	Host               string
	Port               int
	Username           string
	Password           string
	Encryption         EncryptionMode
	From               Address
	AuthIdentity       string
	InsecureSkipVerify bool
}

type Message struct {
	From        *Address
	To          []string
	Cc          []string
	Bcc         []string
	ReplyTo     []string
	Subject     string
	TextBody    string
	HTMLBody    string
	Headers     map[string]string
	Attachments []Attachment
}

type TemplateMessage struct {
	Message
	Template TemplateName
	Data     map[string]any
}

type Attachment struct {
	Filename    string
	ContentType string
	Path        string
	Data        []byte
}

type Service struct {
	cfg      Config
	counter  atomic.Uint64
	randomMu sync.Mutex
	random   *rand.Rand
	sender   smtpSender
}

type smtpSender interface {
	Send(ctx context.Context, account AccountConfig, envelopeFrom string, recipients []string, data []byte, dialTimeout, sendTimeout time.Duration, localName string) error
}

type preparedMessage struct {
	from        *Address
	to          []string
	cc          []string
	bcc         []string
	replyTo     []string
	subject     string
	textBody    string
	htmlBody    string
	headers     map[string]string
	attachments []preparedAttachment
}

type preparedAttachment struct {
	filename    string
	contentType string
	data        []byte
}

type builtBody struct {
	contentType      string
	transferEncoding string
	payload          []byte
}

type sendError struct {
	retryable bool
	message   string
	err       error
}

func (e *sendError) Error() string {
	if e == nil {
		return ""
	}
	if e.message == "" {
		return e.err.Error()
	}
	return e.message + ": " + e.err.Error()
}

func (e *sendError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func New(cfg Config) (*Service, error) {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &Service{
		cfg:    normalized,
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
		sender: smtpClientSender{},
	}, nil
}

func (s *Service) Send(ctx context.Context, msg Message) error {
	if ctx == nil {
		ctx = context.Background()
	}

	prepared, err := s.prepareMessage(msg)
	if err != nil {
		return err
	}

	order := s.accountOrder()
	var lastErr error
	for index, accountIdx := range order {
		account := s.cfg.Accounts[accountIdx]
		from, err := s.resolveFrom(account, prepared.from)
		if err != nil {
			return err
		}

		recipients := envelopeRecipients(prepared)
		payload, err := buildMIMEMessage(prepared, from)
		if err != nil {
			return err
		}

		err = s.sender.Send(ctx, account, from.Email, recipients, payload, s.cfg.DialTimeout, s.cfg.SendTimeout, s.cfg.LocalName)
		if err == nil {
			return nil
		}

		lastErr = fmt.Errorf("send via smtp account %q failed: %w", account.displayName(), err)
		if !isRetryableSendError(err) || index == len(order)-1 {
			return lastErr
		}
	}

	if lastErr == nil {
		lastErr = errors.New("mail send failed")
	}
	return lastErr
}

func (s *Service) SendTemplate(ctx context.Context, msg TemplateMessage) error {
	if msg.Template == "" {
		return errors.New("template is required")
	}
	if strings.TrimSpace(msg.HTMLBody) != "" {
		return errors.New("html_body cannot be used together with template")
	}

	renderedHTML, err := renderTemplate(msg.Template, msg.Data)
	if err != nil {
		return err
	}

	plain := strings.TrimSpace(msg.TextBody)
	message := msg.Message
	message.HTMLBody = renderedHTML
	message.TextBody = plain

	return s.Send(ctx, message)
}

func normalizeConfig(cfg Config) (Config, error) {
	if len(cfg.Accounts) == 0 {
		return Config{}, errors.New("at least one smtp account is required")
	}

	normalized := cfg
	if normalized.DialTimeout <= 0 {
		normalized.DialTimeout = defaultDialTimeout
	}
	if normalized.SendTimeout <= 0 {
		normalized.SendTimeout = defaultSendTimeout
	}
	if normalized.RotationStrategy == "" {
		normalized.RotationStrategy = RotationStrategyRoundRobin
	}
	if !normalized.EnableRotation {
		normalized.RotationStrategy = RotationStrategyNone
	}
	if err := validateRotationStrategy(normalized.RotationStrategy); err != nil {
		return Config{}, err
	}
	if !normalized.DefaultFrom.isZero() {
		if _, err := normalized.DefaultFrom.toMailAddress(); err != nil {
			return Config{}, fmt.Errorf("default_from is invalid: %w", err)
		}
	}

	normalized.Accounts = make([]AccountConfig, len(cfg.Accounts))
	copy(normalized.Accounts, cfg.Accounts)
	for index := range normalized.Accounts {
		account := &normalized.Accounts[index]
		account.Host = strings.TrimSpace(account.Host)
		account.Username = strings.TrimSpace(account.Username)
		account.AuthIdentity = strings.TrimSpace(account.AuthIdentity)
		if account.Port <= 0 {
			return Config{}, fmt.Errorf("smtp account %d port must be greater than 0", index)
		}
		if account.Host == "" {
			return Config{}, fmt.Errorf("smtp account %d host is required", index)
		}
		if account.Encryption == "" {
			account.Encryption = EncryptionSTARTTLS
		}
		switch account.Encryption {
		case EncryptionNone, EncryptionSTARTTLS, EncryptionSSLTLS:
		default:
			return Config{}, fmt.Errorf("smtp account %d encryption %q is unsupported", index, account.Encryption)
		}

		if !account.From.isZero() {
			if _, err := account.From.toMailAddress(); err != nil {
				return Config{}, fmt.Errorf("smtp account %d from is invalid: %w", index, err)
			}
		}
	}

	return normalized, nil
}

func validateRotationStrategy(strategy RotationStrategy) error {
	switch strategy {
	case RotationStrategyNone, RotationStrategyRoundRobin, RotationStrategyRandom:
		return nil
	default:
		return fmt.Errorf("rotation strategy %q is unsupported", strategy)
	}
}

func (s *Service) prepareMessage(msg Message) (preparedMessage, error) {
	to, err := parseAddressSlice(msg.To, "to")
	if err != nil {
		return preparedMessage{}, err
	}
	cc, err := parseAddressSlice(msg.Cc, "cc")
	if err != nil {
		return preparedMessage{}, err
	}
	bcc, err := parseAddressSlice(msg.Bcc, "bcc")
	if err != nil {
		return preparedMessage{}, err
	}
	replyTo, err := parseAddressSlice(msg.ReplyTo, "reply_to")
	if err != nil {
		return preparedMessage{}, err
	}
	if len(to) == 0 && len(cc) == 0 && len(bcc) == 0 {
		return preparedMessage{}, errors.New("at least one recipient is required")
	}

	subject := strings.TrimSpace(msg.Subject)
	if subject == "" {
		return preparedMessage{}, errors.New("subject is required")
	}

	textBody := strings.TrimSpace(msg.TextBody)
	htmlBody := strings.TrimSpace(msg.HTMLBody)
	if textBody == "" && htmlBody == "" {
		return preparedMessage{}, errors.New("either text_body or html_body is required")
	}

	var from *Address
	if msg.From != nil {
		copied := *msg.From
		if _, err := copied.toMailAddress(); err != nil {
			return preparedMessage{}, fmt.Errorf("from is invalid: %w", err)
		}
		from = &copied
	}

	headers, err := sanitizeHeaders(msg.Headers)
	if err != nil {
		return preparedMessage{}, err
	}

	attachments, err := prepareAttachments(msg.Attachments)
	if err != nil {
		return preparedMessage{}, err
	}

	return preparedMessage{
		from:        from,
		to:          to,
		cc:          cc,
		bcc:         bcc,
		replyTo:     replyTo,
		subject:     subject,
		textBody:    textBody,
		htmlBody:    htmlBody,
		headers:     headers,
		attachments: attachments,
	}, nil
}

func prepareAttachments(items []Attachment) ([]preparedAttachment, error) {
	if len(items) == 0 {
		return nil, nil
	}

	prepared := make([]preparedAttachment, 0, len(items))
	for index, item := range items {
		attachment, err := prepareAttachment(item)
		if err != nil {
			return nil, fmt.Errorf("attachment %d is invalid: %w", index, err)
		}
		prepared = append(prepared, attachment)
	}
	return prepared, nil
}

func prepareAttachment(item Attachment) (preparedAttachment, error) {
	var data []byte
	switch {
	case len(item.Data) > 0:
		data = append([]byte(nil), item.Data...)
	case strings.TrimSpace(item.Path) != "":
		raw, err := os.ReadFile(item.Path)
		if err != nil {
			return preparedAttachment{}, fmt.Errorf("read attachment failed: %w", err)
		}
		data = raw
	default:
		return preparedAttachment{}, errors.New("either attachment path or data is required")
	}

	filename := strings.TrimSpace(item.Filename)
	if filename == "" && strings.TrimSpace(item.Path) != "" {
		filename = filepath.Base(item.Path)
	}
	if filename == "" {
		return preparedAttachment{}, errors.New("attachment filename is required")
	}

	contentType := strings.TrimSpace(item.ContentType)
	if contentType == "" {
		contentType = mime.TypeByExtension(strings.ToLower(filepath.Ext(filename)))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return preparedAttachment{
		filename:    filename,
		contentType: contentType,
		data:        data,
	}, nil
}

func sanitizeHeaders(input map[string]string) (map[string]string, error) {
	if len(input) == 0 {
		return nil, nil
	}

	result := make(map[string]string, len(input))
	for key, value := range input {
		trimmedKey := textproto.CanonicalMIMEHeaderKey(strings.TrimSpace(key))
		if trimmedKey == "" {
			return nil, errors.New("header name cannot be empty")
		}
		if isReservedHeader(trimmedKey) {
			return nil, fmt.Errorf("header %q is reserved", trimmedKey)
		}
		if strings.ContainsAny(trimmedKey, "\r\n") {
			return nil, fmt.Errorf("header %q is invalid", trimmedKey)
		}
		result[trimmedKey] = sanitizeHeaderValue(value)
	}
	return result, nil
}

func isReservedHeader(key string) bool {
	switch key {
	case "To", "Cc", "Bcc", "From", "Reply-To", "Subject", "Content-Type", "Mime-Version", "Date":
		return true
	default:
		return false
	}
}

func parseAddressSlice(items []string, field string) ([]string, error) {
	if len(items) == 0 {
		return nil, nil
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		parsed, err := stdmail.ParseAddress(trimmed)
		if err != nil {
			return nil, fmt.Errorf("%s address %q is invalid: %w", field, trimmed, err)
		}
		result = append(result, parsed.String())
	}
	return result, nil
}

func (s *Service) resolveFrom(account AccountConfig, explicit *Address) (Address, error) {
	switch {
	case explicit != nil:
		return *explicit, nil
	case !s.cfg.DefaultFrom.isZero():
		return s.cfg.DefaultFrom, nil
	case !account.From.isZero():
		return account.From, nil
	case account.Username != "":
		address := Address{Email: account.Username}
		if _, err := address.toMailAddress(); err != nil {
			return Address{}, fmt.Errorf("smtp account %q username cannot be used as from address: %w", account.displayName(), err)
		}
		return address, nil
	default:
		return Address{}, fmt.Errorf("smtp account %q has no sender address", account.displayName())
	}
}

func (s *Service) accountOrder() []int {
	total := len(s.cfg.Accounts)
	order := make([]int, 0, total)
	if total == 0 {
		return order
	}

	start := 0
	switch s.cfg.RotationStrategy {
	case RotationStrategyRoundRobin:
		start = int(s.counter.Add(1)-1) % total
	case RotationStrategyRandom:
		s.randomMu.Lock()
		start = s.random.Intn(total)
		s.randomMu.Unlock()
	default:
		start = 0
	}

	for offset := 0; offset < total; offset++ {
		order = append(order, (start+offset)%total)
	}
	return order
}

func envelopeRecipients(msg preparedMessage) []string {
	total := len(msg.to) + len(msg.cc) + len(msg.bcc)
	recipients := make([]string, 0, total)
	recipients = append(recipients, extractEmails(msg.to)...)
	recipients = append(recipients, extractEmails(msg.cc)...)
	recipients = append(recipients, extractEmails(msg.bcc)...)
	return recipients
}

func extractEmails(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		parsed, err := stdmail.ParseAddress(item)
		if err != nil {
			continue
		}
		result = append(result, parsed.Address)
	}
	return result
}

func buildMIMEMessage(msg preparedMessage, from Address) ([]byte, error) {
	var buffer bytes.Buffer

	writeHeader(&buffer, "Date", time.Now().UTC().Format(time.RFC1123Z))
	writeHeader(&buffer, "From", from.String())
	writeHeader(&buffer, "To", strings.Join(msg.to, ", "))
	if len(msg.cc) > 0 {
		writeHeader(&buffer, "Cc", strings.Join(msg.cc, ", "))
	}
	if len(msg.replyTo) > 0 {
		writeHeader(&buffer, "Reply-To", strings.Join(msg.replyTo, ", "))
	}
	writeHeader(&buffer, "Subject", sanitizeHeaderValue(msg.subject))
	writeHeader(&buffer, "MIME-Version", "1.0")

	if len(msg.headers) > 0 {
		keys := make([]string, 0, len(msg.headers))
		for key := range msg.headers {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			writeHeader(&buffer, key, msg.headers[key])
		}
	}

	body, err := buildMessageBody(msg)
	if err != nil {
		return nil, err
	}

	writeHeader(&buffer, "Content-Type", body.contentType)
	if body.transferEncoding != "" {
		writeHeader(&buffer, "Content-Transfer-Encoding", body.transferEncoding)
	}
	buffer.WriteString("\r\n")
	buffer.Write(body.payload)
	return buffer.Bytes(), nil
}

func buildMessageBody(msg preparedMessage) (builtBody, error) {
	if len(msg.attachments) == 0 {
		return buildInlineBody(msg)
	}
	return buildMixedBody(msg)
}

func buildInlineBody(msg preparedMessage) (builtBody, error) {
	if msg.textBody != "" && msg.htmlBody != "" {
		var buffer bytes.Buffer
		writer := multipart.NewWriter(&buffer)

		if err := writeTextPart(writer, "text/plain; charset=UTF-8", msg.textBody); err != nil {
			return builtBody{}, err
		}
		if err := writeTextPart(writer, "text/html; charset=UTF-8", msg.htmlBody); err != nil {
			return builtBody{}, err
		}
		if err := writer.Close(); err != nil {
			return builtBody{}, err
		}

		return builtBody{
			contentType: `multipart/alternative; boundary="` + writer.Boundary() + `"`,
			payload:     buffer.Bytes(),
		}, nil
	}

	contentType := "text/plain; charset=UTF-8"
	body := msg.textBody
	if msg.htmlBody != "" {
		contentType = "text/html; charset=UTF-8"
		body = msg.htmlBody
	}

	var buffer bytes.Buffer
	qp := quotedprintable.NewWriter(&buffer)
	if _, err := qp.Write([]byte(body)); err != nil {
		return builtBody{}, err
	}
	if err := qp.Close(); err != nil {
		return builtBody{}, err
	}
	return builtBody{
		contentType:      contentType,
		transferEncoding: "quoted-printable",
		payload:          buffer.Bytes(),
	}, nil
}

func buildMixedBody(msg preparedMessage) (builtBody, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	if msg.textBody != "" && msg.htmlBody != "" {
		var alternative bytes.Buffer
		altWriter := multipart.NewWriter(&alternative)

		if err := writeTextPart(altWriter, "text/plain; charset=UTF-8", msg.textBody); err != nil {
			return builtBody{}, err
		}
		if err := writeTextPart(altWriter, "text/html; charset=UTF-8", msg.htmlBody); err != nil {
			return builtBody{}, err
		}
		if err := altWriter.Close(); err != nil {
			return builtBody{}, err
		}

		partHeaders := textproto.MIMEHeader{}
		partHeaders.Set("Content-Type", `multipart/alternative; boundary="`+altWriter.Boundary()+`"`)
		part, err := writer.CreatePart(partHeaders)
		if err != nil {
			return builtBody{}, err
		}
		if _, err := part.Write(alternative.Bytes()); err != nil {
			return builtBody{}, err
		}
	} else {
		contentType := "text/plain; charset=UTF-8"
		body := msg.textBody
		if msg.htmlBody != "" {
			contentType = "text/html; charset=UTF-8"
			body = msg.htmlBody
		}
		if err := writeTextPart(writer, contentType, body); err != nil {
			return builtBody{}, err
		}
	}

	for _, attachment := range msg.attachments {
		partHeaders := textproto.MIMEHeader{}
		partHeaders.Set("Content-Type", attachment.contentType)
		partHeaders.Set("Content-Transfer-Encoding", "base64")
		partHeaders.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, escapeHeaderFilename(attachment.filename)))

		part, err := writer.CreatePart(partHeaders)
		if err != nil {
			return builtBody{}, err
		}
		if _, err := io.WriteString(part, wrapBase64(attachment.data)); err != nil {
			return builtBody{}, err
		}
	}

	if err := writer.Close(); err != nil {
		return builtBody{}, err
	}

	return builtBody{
		contentType: `multipart/mixed; boundary="` + writer.Boundary() + `"`,
		payload:     buffer.Bytes(),
	}, nil
}

func writeTextPart(writer *multipart.Writer, contentType, body string) error {
	partHeaders := textproto.MIMEHeader{}
	partHeaders.Set("Content-Type", contentType)
	partHeaders.Set("Content-Transfer-Encoding", "quoted-printable")

	part, err := writer.CreatePart(partHeaders)
	if err != nil {
		return err
	}

	qp := quotedprintable.NewWriter(part)
	if _, err := qp.Write([]byte(body)); err != nil {
		return err
	}
	return qp.Close()
}

func writeHeader(buffer *bytes.Buffer, key, value string) {
	buffer.WriteString(key)
	buffer.WriteString(": ")
	buffer.WriteString(value)
	buffer.WriteString("\r\n")
}

func sanitizeHeaderValue(value string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(strings.TrimSpace(value))
}

func escapeHeaderFilename(name string) string {
	return strings.NewReplacer(`"`, "", "\r", "", "\n", "").Replace(name)
}

func wrapBase64(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	if encoded == "" {
		return ""
	}

	const lineLength = 76
	var builder strings.Builder
	for start := 0; start < len(encoded); start += lineLength {
		end := start + lineLength
		if end > len(encoded) {
			end = len(encoded)
		}
		builder.WriteString(encoded[start:end])
		builder.WriteString("\r\n")
	}
	return builder.String()
}

func renderTemplate(name TemplateName, data map[string]any) (string, error) {
	source, ok := templateSources[name]
	if !ok {
		return "", fmt.Errorf("template %q is unsupported", name)
	}

	tpl, err := template.New(string(name)).Option("missingkey=zero").Parse(source)
	if err != nil {
		return "", fmt.Errorf("parse template %q failed: %w", name, err)
	}

	var buffer bytes.Buffer
	if err := tpl.Execute(&buffer, templateData(data)); err != nil {
		return "", fmt.Errorf("render template %q failed: %w", name, err)
	}
	return buffer.String(), nil
}

func templateData(data map[string]any) map[string]any {
	normalized := map[string]any{
		"app_name":       "Awesome Fiber Template",
		"headline":       "",
		"recipient_name": "",
		"intro":          "",
		"action_url":     "",
		"action_text":    "",
		"code":           "",
		"expires_in":     "",
		"footer":         "",
		"reset_url":      "",
	}
	for key, value := range data {
		normalized[key] = value
	}
	return normalized
}

func isRetryableSendError(err error) bool {
	var classified *sendError
	if errors.As(err, &classified) {
		return classified.retryable
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	return errors.Is(err, io.EOF) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}

func newRetryableError(message string, err error) error {
	return &sendError{retryable: true, message: message, err: err}
}

func newPermanentError(message string, err error) error {
	return &sendError{retryable: false, message: message, err: err}
}

func classifySMTPError(message string, err error, retryOnUnknown bool) error {
	if err == nil {
		return nil
	}

	var textErr *textproto.Error
	if errors.As(err, &textErr) {
		switch {
		case textErr.Code >= 400 && textErr.Code < 500:
			return newRetryableError(message, err)
		case textErr.Code >= 500 && textErr.Code < 600:
			if retryOnUnknown {
				return newRetryableError(message, err)
			}
			return newPermanentError(message, err)
		}
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return newRetryableError(message, err)
	}
	if errors.Is(err, io.EOF) {
		return newRetryableError(message, err)
	}
	if retryOnUnknown {
		return newRetryableError(message, err)
	}
	return newPermanentError(message, err)
}

type smtpClientSender struct{}

func (smtpClientSender) Send(ctx context.Context, account AccountConfig, envelopeFrom string, recipients []string, data []byte, dialTimeout, sendTimeout time.Duration, localName string) error {
	address := net.JoinHostPort(account.Host, strconv.Itoa(account.Port))
	conn, err := (&net.Dialer{Timeout: dialTimeout}).DialContext(ctx, "tcp", address)
	if err != nil {
		return newRetryableError("smtp dial failed", err)
	}

	deadline := connectionDeadline(ctx, sendTimeout, dialTimeout)
	if !deadline.IsZero() {
		_ = conn.SetDeadline(deadline)
	}

	if account.Encryption == EncryptionSSLTLS {
		tlsConn := tls.Client(conn, tlsConfigForAccount(account))
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			_ = conn.Close()
			return newRetryableError("smtp tls handshake failed", err)
		}
		conn = tlsConn
	}

	client, err := smtp.NewClient(conn, account.Host)
	if err != nil {
		_ = conn.Close()
		return newRetryableError("smtp client initialization failed", err)
	}
	defer func() {
		_ = client.Close()
	}()

	if localName != "" {
		if err := client.Hello(localName); err != nil {
			return classifySMTPError("smtp hello failed", err, true)
		}
	}

	if account.Encryption == EncryptionSTARTTLS {
		ok, _ := client.Extension("STARTTLS")
		if !ok {
			return newRetryableError("smtp starttls is not supported by server", errors.New("STARTTLS not supported"))
		}
		if err := client.StartTLS(tlsConfigForAccount(account)); err != nil {
			return classifySMTPError("smtp starttls failed", err, true)
		}
	}

	if account.Username != "" {
		auth := smtp.PlainAuth(account.AuthIdentity, account.Username, account.Password, account.Host)
		if err := client.Auth(auth); err != nil {
			return classifySMTPError("smtp auth failed", err, true)
		}
	}

	if err := client.Mail(envelopeFrom); err != nil {
		return classifySMTPError("smtp mail from failed", err, false)
	}
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return classifySMTPError("smtp rcpt failed", err, false)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return classifySMTPError("smtp data start failed", err, false)
	}
	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		return classifySMTPError("smtp data write failed", err, true)
	}
	if err := writer.Close(); err != nil {
		return classifySMTPError("smtp data close failed", err, false)
	}

	if err := client.Quit(); err != nil {
		return classifySMTPError("smtp quit failed", err, true)
	}
	return nil
}

func connectionDeadline(ctx context.Context, sendTimeout, dialTimeout time.Duration) time.Time {
	timeout := sendTimeout
	if timeout <= 0 {
		timeout = dialTimeout
	}

	var deadline time.Time
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}

	if ctxDeadline, ok := ctx.Deadline(); ok {
		if deadline.IsZero() || ctxDeadline.Before(deadline) {
			deadline = ctxDeadline
		}
	}

	return deadline
}

func tlsConfigForAccount(account AccountConfig) *tls.Config {
	return &tls.Config{
		ServerName:         account.Host,
		InsecureSkipVerify: account.InsecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
	}
}

func (a Address) isZero() bool {
	return strings.TrimSpace(a.Email) == "" && strings.TrimSpace(a.Name) == ""
}

func (a Address) String() string {
	address, err := a.toMailAddress()
	if err != nil {
		return ""
	}
	return address.String()
}

func (a Address) toMailAddress() (*stdmail.Address, error) {
	email := strings.TrimSpace(a.Email)
	if email == "" {
		return nil, errors.New("email is required")
	}
	parsed, err := stdmail.ParseAddress(email)
	if err != nil {
		return nil, err
	}
	return &stdmail.Address{
		Name:    sanitizeHeaderValue(a.Name),
		Address: parsed.Address,
	}, nil
}

func (a AccountConfig) displayName() string {
	if strings.TrimSpace(a.Name) != "" {
		return a.Name
	}
	return net.JoinHostPort(a.Host, strconv.Itoa(a.Port))
}

var templateSources = map[TemplateName]string{
	TemplateWelcome: `<!DOCTYPE html>
<html>
<body style="font-family:Arial,sans-serif;background:#f6f7fb;color:#1f2937;padding:24px;">
  <div style="max-width:640px;margin:0 auto;background:#ffffff;border-radius:16px;padding:32px;">
    <p style="font-size:13px;color:#6b7280;margin:0 0 12px;">{{.app_name}}</p>
    <h1 style="margin:0 0 16px;font-size:28px;">{{if .headline}}{{.headline}}{{else}}Welcome aboard{{end}}</h1>
    <p style="line-height:1.8;">{{if .recipient_name}}Hi {{.recipient_name}},{{else}}Hi,{{end}}</p>
    <p style="line-height:1.8;">{{if .intro}}{{.intro}}{{else}}Your account is ready to go. We are glad to have you here.{{end}}</p>
    {{if .action_url}}<p style="margin:28px 0;"><a href="{{.action_url}}" style="display:inline-block;padding:12px 22px;background:#111827;color:#ffffff;text-decoration:none;border-radius:999px;">{{if .action_text}}{{.action_text}}{{else}}Open now{{end}}</a></p>{{end}}
    {{if .footer}}<p style="margin:24px 0 0;color:#6b7280;font-size:13px;">{{.footer}}</p>{{end}}
  </div>
</body>
</html>`,
	TemplateVerifyCode: `<!DOCTYPE html>
<html>
<body style="font-family:Arial,sans-serif;background:#f6f7fb;color:#1f2937;padding:24px;">
  <div style="max-width:640px;margin:0 auto;background:#ffffff;border-radius:16px;padding:32px;">
    <p style="font-size:13px;color:#6b7280;margin:0 0 12px;">{{.app_name}}</p>
    <h1 style="margin:0 0 16px;font-size:28px;">Verification code</h1>
    <p style="line-height:1.8;">{{if .recipient_name}}Hi {{.recipient_name}},{{else}}Hi,{{end}}</p>
    <p style="line-height:1.8;">Use the code below to complete your verification.</p>
    <div style="margin:28px 0;padding:18px 24px;background:#111827;color:#ffffff;border-radius:14px;font-size:32px;letter-spacing:8px;text-align:center;">{{.code}}</div>
    {{if .expires_in}}<p style="line-height:1.8;">This code will expire in {{.expires_in}}.</p>{{end}}
    {{if .footer}}<p style="margin:24px 0 0;color:#6b7280;font-size:13px;">{{.footer}}</p>{{end}}
  </div>
</body>
</html>`,
	TemplateResetPassword: `<!DOCTYPE html>
<html>
<body style="font-family:Arial,sans-serif;background:#f6f7fb;color:#1f2937;padding:24px;">
  <div style="max-width:640px;margin:0 auto;background:#ffffff;border-radius:16px;padding:32px;">
    <p style="font-size:13px;color:#6b7280;margin:0 0 12px;">{{.app_name}}</p>
    <h1 style="margin:0 0 16px;font-size:28px;">Reset your password</h1>
    <p style="line-height:1.8;">{{if .recipient_name}}Hi {{.recipient_name}},{{else}}Hi,{{end}}</p>
    <p style="line-height:1.8;">We received a request to reset your password.</p>
    {{if .reset_url}}<p style="margin:28px 0;"><a href="{{.reset_url}}" style="display:inline-block;padding:12px 22px;background:#111827;color:#ffffff;text-decoration:none;border-radius:999px;">{{if .action_text}}{{.action_text}}{{else}}Reset password{{end}}</a></p>{{end}}
    {{if .expires_in}}<p style="line-height:1.8;">For security reasons, this link expires in {{.expires_in}}.</p>{{end}}
    {{if .footer}}<p style="margin:24px 0 0;color:#6b7280;font-size:13px;">{{.footer}}</p>{{end}}
  </div>
</body>
</html>`,
}
