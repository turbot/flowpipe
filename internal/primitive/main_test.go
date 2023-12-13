package primitive

import (
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Start MailHog before running tests
	startMailHog()

	time.Sleep(2 * time.Second) // Wait for the server to be ready

	// Run tests
	code := m.Run()

	// Stop MailHog after tests are completed
	stopMailHog()

	// Exit with the test code
	os.Exit(code)
}
