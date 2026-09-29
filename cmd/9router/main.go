package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dresar/go-9router/internal/cli"
	"github.com/dresar/go-9router/internal/config"
	"github.com/dresar/go-9router/internal/handlers"
	"github.com/dresar/go-9router/internal/logging"
	"github.com/dresar/go-9router/internal/middleware"
	"github.com/dresar/go-9router/internal/router"
	"github.com/dresar/go-9router/internal/server"
	"github.com/dresar/go-9router/internal/storage"
	"github.com/dresar/go-9router/internal/updater"
)

const version = "0.5.91"

func main() {
	args := os.Args[1:]

	// Subcommands
	if len(args) >= 2 && args[0] == "xai" && args[1] == "video" {
		if err := cli.RunXaiVideo(args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if len(args) >= 1 {
		switch args[0] {
		case "status":
			runStatus()
			return
		case "stop":
			runStop()
			return
		case "menu":
			cli.RunInteractiveMenu(version, 20128, "127.0.0.1")
			return
		case "version", "-v", "--version":
			fmt.Printf("go-9router version %s\n", version)
			return
		}
	}

	// Flag parsing for server execution
	fs := flag.NewFlagSet("9router", flag.ExitOnError)
	portFlag := fs.String("port", "", "Port to run the server (default: 20128)")
	pShort := fs.String("p", "", "Port (shorthand)")
	directPortFlag := fs.String("direct-port", "", "Direct port without API key (default: 20129)")
	dpShort := fs.String("dp", "", "Direct port (shorthand)")
	noAuthPortsFlag := fs.String("no-auth-ports", "", "Comma-separated ports that bypass API key authentication")
	hostFlag := fs.String("host", "", "Host to bind (default: 0.0.0.0)")
	hShort := fs.String("H", "", "Host (shorthand)")
	noBrowser := fs.Bool("no-browser", false, "Don't open browser automatically")
	nShort := fs.Bool("n", false, "Don't open browser automatically (shorthand)")
	showLog := fs.Bool("log", false, "Show server logs")
	lShort := fs.Bool("l", false, "Show server logs (shorthand)")
	interactive := fs.Bool("interactive", false, "Run in interactive menu mode")
	iShort := fs.Bool("i", false, "Run in interactive menu mode (shorthand)")

	fs.Usage = func() {
		fmt.Printf(`Usage: 9router [command] [options]

Commands:
  xai video --prompt "..."   Generate a Grok Imagine video via running gateway
  status                     Check status and health of running 9router instance
  stop                       Stop running 9router gateway
  menu                       Open interactive terminal control menu

Options:
  -p, --port <port>          Port to run the server (default: 20128)
  -dp, --direct-port <port>  Direct port without API key (default: 20129)
  --no-auth-ports <ports>    Comma-separated ports that bypass API key
  -H, --host <host>          Host to bind (default: 0.0.0.0)
  -n, --no-browser           Don't open browser automatically
  -l, --log                  Show server logs
  -i, --interactive          Run interactive menu
  -v, --version              Show version
  -h, --help                 Show this help message
`)
	}

	_ = fs.Parse(args)

	if *interactive || *iShort {
		cli.RunInteractiveMenu(version, 20128, "127.0.0.1")
		return
	}

	cfg := config.Load()

	// Override config with flags if provided
	portStr := *portFlag
	if portStr == "" {
		portStr = *pShort
	}
	if portStr != "" {
		cfg.Port = portStr
	}

	directPortStr := *directPortFlag
	if directPortStr == "" {
		directPortStr = *dpShort
	}
	if directPortStr != "" {
		cfg.DirectPort = directPortStr
	}
	if *noAuthPortsFlag != "" {
		cfg.NoAuthPorts = *noAuthPortsFlag
	}

	hostStr := *hostFlag
	if hostStr == "" {
		hostStr = *hShort
	}
	if hostStr == "" {
		hostStr = "0.0.0.0"
	}

	dontOpenBrowser := *noBrowser || *nShort
	_ = *showLog || *lShort

	portNum, err := strconv.Atoi(cfg.Port)
	if err != nil {
		portNum = 20128
	}

	directPortNum, err := strconv.Atoi(cfg.DirectPort)
	if err != nil || directPortNum <= 0 {
		directPortNum = 20129
	}

	// Clean up stale processes on the ports
	cli.KillProcessOnPort(portNum)
	if cfg.EnableDirectPort && directPortNum > 0 && directPortNum != portNum {
		cli.KillProcessOnPort(directPortNum)
	}

	cli.ShowBanner(version)
	displayHost := hostStr
	if displayHost == "0.0.0.0" {
		displayHost = "localhost"
	}
	lanIP := cli.GetLanIP()
	if lanIP != "" && hostStr == "0.0.0.0" {
		fmt.Printf("🌐 LAN Address:       http://%s:%d\n", lanIP, portNum)
	}
	fmt.Printf("🚀 Web Dashboard:     http://%s:%d/endpoint\n", displayHost, portNum)
	fmt.Printf("🔌 API Endpoint:      http://%s:%d/v1 (Protected/Auth)\n", displayHost, portNum)
	if cfg.EnableDirectPort && directPortNum > 0 {
		fmt.Printf("⚡ Direct Endpoint:   http://%s:%d/v1 (No API Key Required)\n", displayHost, directPortNum)
		if lanIP != "" && hostStr == "0.0.0.0" {
			fmt.Printf("⚡ Direct LAN:        http://%s:%d/v1 (No API Key Required)\n", lanIP, directPortNum)
		}
	}
	fmt.Println()

	db, err := storage.Open(cfg.DataDir)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer storage.Close()

	if err := storage.Migrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	h := &handlers.Handler{DB: db, Cfg: cfg}
	mux := router.New(h, cfg.JWTSecret)

	// Main server (port e.g. 20128) - with ServerPortKey = portNum, IsDirectNoAuth = false
	addr := fmt.Sprintf("%s:%d", hostStr, portNum)
	mainHandler := middleware.WithServerPort(portNum, false)(mux)
	srv := server.New(addr, mainHandler, cfg.ReadHeaderTimeout, cfg.IdleTimeout)

	go func() {
		if err := srv.Start(); err != nil && err.Error() != "http: Server closed" {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Direct server without API key (port e.g. 20129) - with ServerPortKey = directPortNum, IsDirectNoAuth = true
	var directSrv *server.Server
	if cfg.EnableDirectPort && directPortNum > 0 && directPortNum != portNum {
		directAddr := fmt.Sprintf("%s:%d", hostStr, directPortNum)
		directHandler := middleware.WithServerPort(directPortNum, true)(mux)
		directSrv = server.New(directAddr, directHandler, cfg.ReadHeaderTimeout, cfg.IdleTimeout)

		go func() {
			if err := directSrv.Start(); err != nil && err.Error() != "http: Server closed" {
				logging.Warn("SERVER", "direct server error", "err", err)
			}
		}()
	}

	// Start background 1-hour periodic GitHub sync check
	updater.GetManager().StartScheduler(1 * time.Hour)

	// Auto open browser
	if !dontOpenBrowser {
		go func() {
			if cli.WaitServerReady(portNum, 5*time.Second) {
				url := fmt.Sprintf("http://%s:%d/endpoint", displayHost, portNum)
				_ = cli.OpenBrowser(url)
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logging.Info("SERVER", "shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	if directSrv != nil {
		_ = directSrv.Shutdown(ctx)
	}
	logging.Info("SERVER", "stopped gracefully")
}

func runStatus() {
	client := cli.NewAPIClient("127.0.0.1", 20128)
	health, err := client.GetHealth()
	if err != nil {
		fmt.Println("❌ 9Router is NOT running (or port 20128 is unreachable)")
		return
	}
	fmt.Println("✅ 9Router is RUNNING")
	for k, v := range health {
		fmt.Printf("   %s: %v\n", strings.Title(k), v)
	}
}

func runStop() {
	fmt.Println("🛑 Stopping 9Router...")
	cli.KillProcessOnPort(20128)
	client := cli.NewAPIClient("127.0.0.1", 20128)
	_ = client.Shutdown()
	fmt.Println("✅ 9Router stopped.")
}
