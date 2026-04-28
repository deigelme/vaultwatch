package alert_test

import (
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

// startFakeSMTP starts a minimal TCP server that accepts one SMTP session
// and returns the raw data it received.
func startFakeSMTP(t *testing.T) (addr string, received <-chan string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start fake SMTP: %v", err)
	}

	ch := make(chan string, 1)
	go func() {
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(3 * time.Second)) //nolint:errcheck

		// Speak just enough SMTP to satisfy net/smtp.
		fmt.Fprintf(conn, "220 fake SMTP ready\r\n")
		buf := make([]byte, 4096)
		var sb strings.Builder
		for {
			n, err := conn.Read(buf)
			if n > 0 {
				sb.Write(buf[:n])
				data := sb.String()
				if strings.Contains(data, "EHLO") {
					fmt.Fprintf(conn, "250 OK\r\n")
				}
				if strings.Contains(data, "MAIL FROM") {
					fmt.Fprintf(conn, "250 OK\r\n")
				}
				if strings.Contains(data, "RCPT TO") {
					fmt.Fprintf(conn, "250 OK\r\n")
				}
				if strings.Contains(data, "DATA") && !strings.Contains(data, "\r\n.\r\n") {
					fmt.Fprintf(conn, "354 Start input\r\n")
				}
				if strings.Contains(data, "\r\n.\r\n") {
					fmt.Fprintf(conn, "250 OK\r\n")
				}
				if strings.Contains(data, "QUIT") {
					fmt.Fprintf(conn, "221 Bye\r\n")
					break
				}
			}
			if err == io.EOF || err != nil {
				break
			}
		}
		ch <- sb.String()
	}()

	return ln.Addr().String(), ch
}

func TestEmailNotifier_Send(t *testing.T) {
	addr, received := startFakeSMTP(t)
	host, portStr, _ := net.SplitHostPort(addr)
	port := 0
	fmt.Sscanf(portStr, "%d", &port)

	cfg := alert.EmailConfig{
		SMTPHost: host,
		SMTPPort: port,
		From:     "vaultwatch@example.com",
		To:       []string{"ops@example.com"},
	}
	notifier := alert.NewEmailNotifier(cfg)

	a := alert.Alert{
		Level:      alert.LevelWarning,
		SecretPath: "secret/db/password",
		Message:    "expires in 48h",
	}

	if err := notifier.Send(a); err != nil {
		t.Fatalf("Send() returned unexpected error: %v", err)
	}

	select {
	case data := <-received:
		if !strings.Contains(data, "secret/db/password") {
			t.Errorf("expected secret path in email body, got:\n%s", data)
		}
		if !strings.Contains(data, "WARNING") {
			t.Errorf("expected alert level in email body, got:\n%s", data)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for fake SMTP to receive message")
	}
}
