//go:build windows

package mpv

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"time"

	"github.com/Microsoft/go-winio"
)

func createMpv() (*mpv, error) {
	pipe := `\\.\pipe\jfsh-mpv-` + strconv.FormatInt(time.Now().UnixNano(), 10)
	cmd := exec.Command("mpv", "--idle", "--input-ipc-server="+pipe)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to create mpv: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wait for socket to be created
	var conn net.Conn
	var err error
	for range 300 {
		conn, err = winio.DialPipeContext(ctx, pipe)
		if err == nil {
			break
		}
		if err := ctx.Err(); err != nil {
			cmd.Process.Kill()
			return nil, fmt.Errorf("failed to connect to mpv socket: %w", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to connect to mpv socket: %w", err)
	}
	return &mpv{
		conn:    conn,
		scanner: bufio.NewScanner(conn),
		cmd:     cmd,
		socket:  pipe,
		oldMpv:  isOldMpv(),
	}, nil
}
