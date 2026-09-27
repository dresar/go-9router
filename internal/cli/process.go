package cli

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func WaitServerReady(port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 200*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(150 * time.Millisecond)
	}
	return false
}

func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func KillProcessOnPort(port int) {
	if runtime.GOOS == "windows" {
		cmdStr := fmt.Sprintf(`netstat -ano | findstr :%d`, port)
		out, err := exec.Command("cmd", "/C", cmdStr).Output()
		if err != nil {
			return
		}
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "LISTENING") {
				fields := strings.Fields(line)
				if len(fields) > 0 {
					pid := fields[len(fields)-1]
					_ = exec.Command("taskkill", "/F", "/PID", pid).Run()
				}
			}
		}
	} else {
		out, err := exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port)).Output()
		if err == nil && len(out) > 0 {
			pid := strings.TrimSpace(string(out))
			_ = exec.Command("kill", "-9", pid).Run()
		}
	}
	time.Sleep(300 * time.Millisecond)
}

func KillByPIDFile(pidFile string) {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return
	}
	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil || pid <= 0 {
		_ = os.Remove(pidFile)
		return
	}

	if runtime.GOOS == "windows" {
		_ = exec.Command("taskkill", "/F", "/T", "/PID", pidStr).Run()
	} else {
		p, err := os.FindProcess(pid)
		if err == nil {
			_ = p.Kill()
		}
	}
	_ = os.Remove(pidFile)
}

func GetAppDataDir() string {
	if dir := os.Getenv("DATA_DIR"); dir != "" {
		return dir
	}
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, _ := os.UserHomeDir()
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "9router")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".9router")
}

func GetLanIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
