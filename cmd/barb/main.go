package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/travisbale/barb/internal/app"
	"github.com/travisbale/barb/internal/crypto"
	"github.com/travisbale/barb/internal/crypto/aes"
	"github.com/travisbale/barb/internal/delivery"
	"github.com/travisbale/barb/internal/store/sqlite"
)

//go:embed all:dist
var frontendFS embed.FS

var Version = "dev"

func main() {
	var (
		addr    string
		dbPath  string
		tlsCert string
		tlsKey  string
		debug   bool
	)

	root := &cobra.Command{
		Use:          "barb",
		Short:        "Barb — campaign management console for Mirage",
		RunE:         func(cmd *cobra.Command, args []string) error { return cmd.Help() },
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVar(&addr, "addr", ":443", "listen address")
	root.PersistentFlags().StringVar(&dbPath, "db", "barb.db", "SQLite database path")
	root.PersistentFlags().StringVar(&tlsCert, "tls-cert", "", "TLS certificate path (auto-generated if empty)")
	root.PersistentFlags().StringVar(&tlsKey, "tls-key", "", "TLS key path (auto-generated if empty)")
	root.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the console server (default)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cmd.Context(), addr, dbPath, tlsCert, tlsKey, debug)
		},
	}
	root.RunE = serveCmd.RunE

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version and exit",
		Run:   func(cmd *cobra.Command, args []string) { fmt.Println(Version) },
	}

	root.AddCommand(serveCmd, versionCmd)

	if err := run(root); err != nil {
		os.Exit(1)
	}
}

func run(root *cobra.Command) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	return root.ExecuteContext(ctx)
}

func runServe(ctx context.Context, addr, dbPath, tlsCert, tlsKey string, debug bool) error {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	db, err := sqlite.Open(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	keyPath := filepath.Join(filepath.Dir(dbPath), "encryption.key")
	encryptionKey, err := crypto.LoadOrGenerateKey(keyPath)
	if err != nil {
		return err
	}
	enc := aes.NewCipher(encryptionKey)

	frontendDist, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		return fmt.Errorf("loading embedded frontend: %w", err)
	}

	application := app.New(app.Config{
		DB:       db,
		Cipher:   enc,
		Frontend: frontendDist,
		Mailer:   &delivery.Sender{Logger: logger},
		Version:  Version,
		Logger:   logger,
	})

	// Set up TLS certificate.
	dataDir := filepath.Dir(dbPath)
	if tlsCert == "" {
		tlsCert = filepath.Join(dataDir, "tls.crt")
	}
	if tlsKey == "" {
		tlsKey = filepath.Join(dataDir, "tls.key")
	}
	if err := crypto.LoadOrGenerateTLS(tlsCert, tlsKey); err != nil {
		return fmt.Errorf("setting up TLS: %w", err)
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: application.Handler(),
	}

	go func() {
		logger.Info("barb starting", "addr", addr, "version", Version, "tls_cert", tlsCert)
		if err := srv.ListenAndServeTLS(tlsCert, tlsKey); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	application.Shutdown()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
	}

	return nil
}
