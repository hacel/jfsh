//go:build !windows

package mpv

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func createMpv() (*mpv, error) {
	socket := filepath.Join(os.TempDir(), fmt.Sprintf("jfsh-mpv-socket-%d", time.Now().UnixNano()))
	cmd := exec.Command("mpv", "--idle", "--input-ipc-server="+socket)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to create mpv: %w", err)
	}
	slog.Debug("created mpv", "socket", socket)

	// Wait for socket to be created
	var conn net.Conn
	var err error
	for range 300 {
		conn, err = net.Dial("unix", socket)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to connect to mpv socket: %w", err)
	}
	slog.Debug("successfully connected to mpv", "socket", socket)
	return &mpv{
		conn:    conn,
		scanner: bufio.NewScanner(conn),
		cmd:     cmd,
		socket:  socket,
	}, nil
}
