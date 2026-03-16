package ipc

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// socketPath returns the Unix socket path used for single-instance IPC.
func socketPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	return filepath.Join(home, ".config", "chekout", "chekout.sock"), nil
}

// Send dials the running instance and sends a URL. Returns an error if no
// instance is listening (caller should handle the URL directly).
func Send(rawURL string) error {
	sockPath, err := socketPath()
	if err != nil {
		return err
	}

	conn, err := net.DialTimeout("unix", sockPath, 2*time.Second)
	if err != nil {
		return fmt.Errorf("no running instance: %w", err)
	}
	defer conn.Close()

	fmt.Printf("ipc: relaying URL to running instance: %s\n", rawURL)
	_, err = fmt.Fprintln(conn, rawURL)
	return err
}

// Serve removes any stale socket, listens for connections, and forwards
// received URLs to ch. Intended to run in a goroutine.
func Serve(ch chan<- string) {
	sockPath, err := socketPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ipc: socket path error: %v\n", err)
		return
	}

	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(sockPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "ipc: mkdir: %v\n", err)
		return
	}

	// Remove stale socket from a previous run.
	_ = os.Remove(sockPath)

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ipc: listen: %v\n", err)
		return
	}
	defer ln.Close()

	fmt.Printf("ipc: listening on %s\n", sockPath)

	for {
		conn, err := ln.Accept()
		if err != nil {
			// Listener closed — shut down gracefully.
			return
		}
		fmt.Println("ipc: connection received")
		go handleConn(conn, ch)
	}
}

func handleConn(conn net.Conn, ch chan<- string) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			fmt.Printf("ipc: received URL: %s\n", line)
			ch <- line
		}
	}
}
