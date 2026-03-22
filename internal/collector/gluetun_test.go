package collector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeProcNetDev writes the given content to a temp file and returns its path.
func writeProcNetDev(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "proc_net_dev")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}
	return f.Name()
}

const validProcNetDev = `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo:   12345      100    0    0    0     0          0         0    12345      100    0    0    0     0       0          0
  tun0: 1000000     5000   10   20    0     0          0         0   500000     2500    5   15    0     0       0          0
`

func TestReadInterfaceStats_HappyPath(t *testing.T) {
	path := writeProcNetDev(t, validProcNetDev)

	stats, err := readInterfaceStats("tun0", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.RxBytes != 1000000 {
		t.Errorf("RxBytes: got %d, want 1000000", stats.RxBytes)
	}
	if stats.RxPackets != 5000 {
		t.Errorf("RxPackets: got %d, want 5000", stats.RxPackets)
	}
	if stats.RxErrors != 10 {
		t.Errorf("RxErrors: got %d, want 10", stats.RxErrors)
	}
	if stats.RxDropped != 20 {
		t.Errorf("RxDropped: got %d, want 20", stats.RxDropped)
	}
	if stats.TxBytes != 500000 {
		t.Errorf("TxBytes: got %d, want 500000", stats.TxBytes)
	}
	if stats.TxPackets != 2500 {
		t.Errorf("TxPackets: got %d, want 2500", stats.TxPackets)
	}
	if stats.TxErrors != 5 {
		t.Errorf("TxErrors: got %d, want 5", stats.TxErrors)
	}
	if stats.TxDropped != 15 {
		t.Errorf("TxDropped: got %d, want 15", stats.TxDropped)
	}
}

func TestReadInterfaceStats_InterfaceNotFound(t *testing.T) {
	path := writeProcNetDev(t, validProcNetDev)

	_, err := readInterfaceStats("wg0", path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "wg0") {
		t.Errorf("error %q does not mention interface name", err.Error())
	}
}

func TestReadInterfaceStats_MalformedLine(t *testing.T) {
	content := `Inter-|   Receive
 face |bytes
  tun0: 100 200 300
`
	path := writeProcNetDev(t, content)

	_, err := readInterfaceStats("tun0", path)
	if err == nil {
		t.Fatal("expected error for malformed line, got nil")
	}
}

func TestReadInterfaceStats_FileNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent")

	_, err := readInterfaceStats("tun0", path)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
