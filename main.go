package main

import (
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/energye/systray"
)

//go:embed meow.ico
var icon []byte

func runShutdown() {
	var c *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		c = exec.Command("cmd", "/C", "shutdown", "/s")

	default:
		c = exec.Command("notify-send", "Meow")
	}

	if err := c.Run(); err != nil {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println("Shutdown initiated")
	}
}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("windowsDowner")
	systray.SetIcon(icon)
	systray.SetTooltip("windowsDowner - Remote Shutdown Listener")

	listener, err := net.Listen("tcp", ":1337")
	if err != nil {
		panic(fmt.Sprintf("Failed to listen: %v", err))
	}

	ip := getLocalIP()
	addr := fmt.Sprintf("%s:%d", ip, listener.Addr().(*net.TCPAddr).Port)

	mInfo := systray.AddMenuItem("Listening on "+addr, "Current listening address")
	mInfo.Disable()

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Exit the application")

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			runShutdown()
		})
		if err := http.Serve(listener, nil); err != nil {
			fmt.Println("HTTP server error:", err)
		}
	}()

	mQuit.Click(func() {
		systray.Quit()
	})
}

func onExit() {
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}
	return "127.0.0.1"
}
