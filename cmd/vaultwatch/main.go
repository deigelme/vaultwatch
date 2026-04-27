// Package main is the entry point for the vaultwatch CLI tool.
// It wires together configuration loading, Vault client initialization,
// and the secret expiration monitor.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/vaultwatch/internal/config"
	"github.com/yourusername/vaultwatch/internal/vault"
)

const defaultConfigPath = "configs/vaultwatch.yaml"

func main() {
	configPath := flag.String("config", defaultConfigPath, "path to vaultwatch config file")
	verbose := flag.Bool("verbose", false, "enable verbose logging")
	flag.Parse()

	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	client, err := vault.NewClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating vault client: %v\n", err)
		os.Exit(1)
	}

	monitor := vault.NewMonitor(client, cfg)

	// Set up context that cancels on SIGINT or SIGTERM.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Printf("vaultwatch started — monitoring %d secret path(s) every %s",
		len(cfg.SecretPaths), cfg.CheckInterval)

	if err := monitor.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "monitor exited with error: %v\n", err)
		os.Exit(1)
	}

	log.Println("vaultwatch stopped")
}
