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

	"github.com/spf13/cobra"

	"github.com/travisbale/mirador/internal/api"
	"github.com/travisbale/mirador/internal/delivery"
	"github.com/travisbale/mirador/internal/phishing"
	"github.com/travisbale/mirador/internal/server"
	"github.com/travisbale/mirador/internal/store/sqlite"
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
		Use:          "mirador",
		Short:        "Mirador — campaign management console for Mirage",
		RunE:         func(cmd *cobra.Command, args []string) error { return cmd.Help() },
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVar(&addr, "addr", ":8080", "listen address")
	root.PersistentFlags().StringVar(&dbPath, "db", "mirador.db", "SQLite database path")
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

	frontendDist, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		return fmt.Errorf("loading embedded frontend: %w", err)
	}

	targetStore := sqlite.NewTargetStore(db)
	templateStore := sqlite.NewTemplateStore(db)
	smtpStore := sqlite.NewSMTPStore(db)

	targetSvc := &phishing.TargetService{Store: targetStore}
	templateSvc := &phishing.TemplateService{Store: templateStore}
	smtpSvc := &phishing.SMTPService{Store: smtpStore}
	miragedSvc := &phishing.MiragedService{Store: sqlite.NewMiragedStore(db)}

	monitor := &phishing.SessionMonitor{
		Campaigns: sqlite.NewCampaignStore(db),
		Miraged:   miragedSvc,
		Logger:    logger,
	}

	campaignSvc := &phishing.CampaignService{
		Store:     sqlite.NewCampaignStore(db),
		Targets:   targetStore,
		Templates: templateStore,
		SMTP:      smtpStore,
		Miraged:   miragedSvc,
		Monitor:   monitor,
		Mailer:    &delivery.Sender{Logger: logger},
		Logger:    logger,
	}

	apiRouter := &api.Router{
		Miraged:   miragedSvc,
		Campaigns: campaignSvc,
		Targets:   targetSvc,
		Templates: templateSvc,
		SMTP:      smtpSvc,
		Version:   Version,
		Logger:    logger,
	}

	srv := server.New(server.Config{
		Addr:     addr,
		API:      apiRouter,
		Frontend: frontendDist,
	})

	go func() {
		logger.Info("mirador starting", "addr", addr, "version", Version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
	}

	return nil
}
