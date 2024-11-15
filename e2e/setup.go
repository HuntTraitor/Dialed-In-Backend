package e2e

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

const (
	baseBinName = "temp-testbinary"
)

// LaunchTestProgram launches the testing server from a built binary
func LaunchTestProgram(port string) (cleanup func(), sendInterrupt func() error, err error) {
	binName, err := buildBinary()
	if err != nil {
		return nil, nil, err
	}

	sendInterrupt, kill, err := runServer(binName, port)

	cleanup = func() {
		if kill != nil {
			kill()
		}
		os.Remove(binName)
	}
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	return cleanup, sendInterrupt, nil
}

// runServer runs a server from a bin
func runServer(binName string, port string) (sendInterrupt func() error, kill func(), err error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}

	cmdPath := filepath.Join(dir, binName)

	if err := makeExecutable(cmdPath); err != nil {
		fmt.Println("Failed to make binary executable:", err)
		os.Exit(1)
	}

	cmd := exec.Command(cmdPath, "-env=test", "-db-dsn=postgres://postgres:postgres@localhost:5432/test?sslmode=disable", "-port="+port)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("cannot run temp converter: %s", err)
	}

	kill = func() {
		_ = cmd.Process.Kill()
	}

	sendInterrupt = func() error {
		return cmd.Process.Signal(syscall.SIGTERM)
	}

	err = waitForServerListening(port)
	return

}

// waitForServerListening pings the location to confirm a server is listening
func waitForServerListening(port string) error {
	for i := 0; i < 30; i++ {
		conn, _ := net.Dial("tcp", net.JoinHostPort("localhost", port))
		if conn != nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("nothing seems to be listening on localhost:%s", port)
}

// buildBinary builds a binary with a randomString-basebinname
func buildBinary() (string, error) {
	binName := randomString(10) + "-" + baseBinName

	// Prepare the build command
	build := exec.Command("go", "build", "-o", binName, "../cmd/api/.")

	// Capture the output of the build command
	output, err := build.CombinedOutput()
	if err != nil {
		// Print the output to help with debugging
		return "", fmt.Errorf("cannot build tool %s: %s\n%s", binName, err, string(output))
	}
	return binName, nil
}

// randomString takes a number and outputs a random string
func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func makeExecutable(filePath string) error {
	return os.Chmod(filePath, 0755) // Grant execute permissions
}
