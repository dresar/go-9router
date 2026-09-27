package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ShowBanner(version string) {
	fmt.Println()
	fmt.Println("\033[1;36m====================================================\033[0m")
	fmt.Printf("\033[1;32m  🚀 Go 9Router v%s — High Performance AI Gateway\033[0m\n", version)
	fmt.Println("\033[1;36m====================================================\033[0m")
	fmt.Println()
}

func RunInteractiveMenu(version string, port int, host string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		ShowBanner(version)
		displayHost := host
		if displayHost == "0.0.0.0" {
			displayHost = "localhost"
		}
		webURL := fmt.Sprintf("http://%s:%d/dashboard", displayHost, port)
		fmt.Printf("📍 Gateway Endpoint: \033[1;33m%s\033[0m\n\n", webURL)

		fmt.Println("Choose an option:")
		fmt.Println("  [1] 🌐 Web UI (Open in Browser)")
		fmt.Println("  [2] 🔍 Gateway Status & Health Check")
		fmt.Println("  [3] 🎬 Generate Grok Video (xAI Imagine)")
		fmt.Println("  [4] 🛑 Stop Running Gateway")
		fmt.Println("  [5] 🚪 Exit")
		fmt.Print("\nEnter choice [1-5]: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		choice := strings.TrimSpace(input)

		switch choice {
		case "1":
			fmt.Printf("Opening %s in browser...\n", webURL)
			_ = OpenBrowser(webURL)
		case "2":
			client := NewAPIClient(displayHost, port)
			health, err := client.GetHealth()
			if err != nil {
				fmt.Printf("\n❌ Gateway check failed: %v (is it running?)\n", err)
			} else {
				fmt.Printf("\n✅ Gateway is HEALTHY: %+v\n", health)
			}
		case "3":
			fmt.Print("\nEnter video prompt: ")
			prompt, _ := reader.ReadString('\n')
			prompt = strings.TrimSpace(prompt)
			if prompt != "" {
				_ = RunXaiVideo([]string{"--prompt", prompt, "--port", fmt.Sprintf("%d", port), "--host", displayHost})
			}
		case "4":
			fmt.Println("Stopping gateway processes...")
			KillProcessOnPort(port)
			fmt.Println("✅ Done.")
		case "5", "q", "exit":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid choice. Please select 1-5.")
		}

		fmt.Println("\nPress Enter to continue...")
		_, _ = reader.ReadString('\n')
	}
}
