package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/hunttraitor/dialed-in-backend/internal/mailer"
	"github.com/hunttraitor/dialed-in-backend/internal/s3"
	"github.com/hunttraitor/dialed-in-backend/internal/vcs"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var (
	version = vcs.Version()
)

type config struct {
	port    int
	env     string
	metrics bool
	db      struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}

	limiter struct {
		rps        float64
		burst      int
		enabled    bool
		expiration time.Duration
	}

	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}

	s3 struct {
		bucket    string
		region    string
		accessKey string
		secretKey string
	}
}

type application struct {
	config config
	logger *slog.Logger
	models data.Models
	mailer Mailer
	wg     sync.WaitGroup
	s3     *s3.S3
}

type Mailer interface {
	Send(recipient, templateFile string, data any) error
}

func main() {
	var cfg config
	databaseURL := os.Getenv("DATABASE_URL")
	SMTPUsername := os.Getenv("SMTP_USERNAME")
	SMTPPassword := os.Getenv("SMTP_PASSWORD")
	S3Bucket := os.Getenv("S3_BUCKET")
	S3AccessKey := os.Getenv("S3_ACCESS_KEY")
	S3SecretKey := os.Getenv("S3_SECRET_KEY")

	// Setting flags for all the different configurations
	flag.IntVar(&cfg.port, "port", 3000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", databaseURL, "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection ide time")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.DurationVar(&cfg.limiter.expiration, "limiter-expiration", 3*time.Minute, "Set limiter expiration")
	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", SMTPUsername, "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", SMTPPassword, "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Dialed-In <no-reply@dialedincafe.com>", "SMTP sender")
	flag.BoolVar(&cfg.metrics, "metrics", false, "Enable application metrics")
	flag.StringVar(&cfg.s3.bucket, "s3-bucket", S3Bucket, "AWS S3 bucket")
	flag.StringVar(&cfg.s3.region, "s3-region", "us-east-2", "AWS S3 region")
	flag.StringVar(&cfg.s3.accessKey, "s3-access-key", S3AccessKey, "AWS S3 access key")
	flag.StringVar(&cfg.s3.secretKey, "s3-secret-key", S3SecretKey, "AWS S3 secret key")
	displayVersion := flag.Bool("version", false, "Display version and exit")
	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDb(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	logger.Info("Database pool has been established")

	// Add system debug logs to /debug/vars
	expvar.NewString("version").Set(version)

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))

	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	s3Config, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.s3.accessKey, cfg.s3.secretKey, "")),
		awsConfig.WithRegion(cfg.s3.region),
	)

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("AWS S3 Config has been established")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
		s3:     NewS3(s3Config),
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

// openDb returns a connection to the database
func openDb(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set pool configurations
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxLifetime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// NewS3 creates a new S3 object
func NewS3(cfg aws.Config) *s3.S3 {
	client := awsS3.NewFromConfig(cfg)
	presigner := awsS3.NewPresignClient(client)

	return &s3.S3{
		Client:    client,
		Presigner: presigner,
	}
}
