package e2e

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	baseBinName = "temp-testbinary"
)

// LaunchTestProgram launches the testing server from a built binary
func LaunchTestProgram(port string) (cleanup func(), sendInterrupt func() error, err error) {

	// Suppress output to reduce clutter in CI
	suppressOutput()

	err = godotenv.Load("../.env")
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		os.Exit(1)
	}

	binName, err := buildBinary()
	if err != nil {
		fmt.Println("Failed to build binary:", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		fmt.Println("Failed to connect to postgres:", err)
		os.Exit(1)
	}

	if err = goose.Up(db, "../db/migrations"); err != nil {
		db.Close()
		fmt.Println("Failed to up migrations:", err)
		os.Exit(1)
	}

	err = runSeeds(os.Getenv("TEST_DATABASE_URL"), "../db/sql/seeds.sql")
	if err != nil {
		db.Close()
		fmt.Println("Failed to run seeds:", err)
		os.Exit(1)
	}

	sendInterrupt, kill, err := runServer(binName, port)

	cleanup = func() {
		if kill != nil {
			kill()
		}
		if resetErr := goose.Reset(db, "../db/migrations"); resetErr != nil {
			fmt.Println("Failed to reset migrations during cleanup:", resetErr)
		}
		db.Close()
		os.Remove(binName)
		fmt.Println("Cleaned up")
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

	if err = makeExecutable(cmdPath); err != nil {
		fmt.Println("Failed to make binary executable:", err)
		os.Exit(1)
	}

	cmd := exec.Command(
		cmdPath,
		"-env=test",
		"-db-dsn="+os.Getenv("TEST_DATABASE_URL"),
		"-port="+port,
		"-smtp-host=localhost",
		"-smtp-port=1025",
		"-smtp-username="+"",
		"-smtp-password="+"",
		"-limiter-enabled=false",
		"-metrics=true",
	)

	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
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

// buildBinary builds a binary with a randomString-baseBinName
func buildBinary() (string, error) {
	binName := randomString(10) + "-" + baseBinName

	// Prepare the build command
	build := exec.Command("go", "build", "-o", binName, "-buildvcs=false", "../cmd/api/.")

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

func runSeeds(databaseURL string, seedsFilePath string) error {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %w", err)
	}
	defer db.Close()

	seedData, err := os.ReadFile(seedsFilePath)
	if err != nil {
		return fmt.Errorf("failed to read seeds file: %w", err)
	}

	seedStatements := strings.Split(string(seedData), ";")

	for _, stmt := range seedStatements {
		trimmedStmt := strings.TrimSpace(stmt)
		if trimmedStmt == "" {
			continue
		}

		_, err = db.Exec(trimmedStmt)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %s, error: %w", trimmedStmt, err)
		}
	}

	log.Println("Seeding completed successfully.")
	return nil
}

func suppressOutput() {
	// Open /dev/null to suppress output
	devNull, err := os.OpenFile("/dev/null", os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	os.Stdout = devNull
	os.Stderr = devNull

	// Suppress logging by replacing the default logger
	log.SetOutput(devNull)
}
