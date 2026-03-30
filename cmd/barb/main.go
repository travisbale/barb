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
		addr   string
		dbPath string
		debug  bool
	)

	root := &cobra.Command{
		Use:          "barb",
		Short:        "Barb — campaign management console for Mirage",
		RunE:         func(cmd *cobra.Command, args []string) error { return cmd.Help() },
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVar(&addr, "addr", ":8080", "listen address")
	root.PersistentFlags().StringVar(&dbPath, "db", "barb.db", "SQLite database path")
	root.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the console server (default)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cmd.Context(), addr, dbPath, debug)
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

func runServe(ctx context.Context, addr, dbPath string, debug bool) error {
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

	srv := &http.Server{
		Addr:    addr,
		Handler: application.Handler(),
	}

	go func() {
		logger.Info("barb starting", "addr", addr, "version", Version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
