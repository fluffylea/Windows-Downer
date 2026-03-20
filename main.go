package main

import (
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/energye/systray"
)

//go:embed meow.ico
var icon []byte

func runShutdown(duration time.Duration) {
	var c *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		c = exec.Command("cmd", "/C", "shutdown", "/s", "/t", strconv.Itoa(int(duration.Seconds())))
	case "darwin":
		s := fmt.Sprintf("display notification \"%s%s\" with title \"%s\"", "Shutdown in ", duration.String(), "Shutdown")
		c = exec.Command("osascript", "-e", s)
	default:
		c = exec.Command("notify-send", fmt.Sprintf("Shutdown in %s", duration.String()))
	}

	if err := c.Run(); err != nil {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println("Shutdown initiated")
	}
}

func cancelShutdown() {
	var c *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		c = exec.Command("cmd", "/C", "shutdown", "/a")
	case "darwin":
		s := fmt.Sprintf("display notification \"%s\" with title \"%s\"", "Shutdown cancelled", "Shutdown")
		c = exec.Command("osascript", "-e", s)
	default:
		c = exec.Command("notify-send", fmt.Sprintf("Shutdown cancelled"))
	}

	if err := c.Run(); err != nil {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println("Shutdown cancelled")
	}
}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("")
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
		http.HandleFunc("/cancel", func(w http.ResponseWriter, r *http.Request) {
			cancelShutdown()
		})

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			runShutdown(60 * time.Second)
		})

		http.HandleFunc("/{duration}", func(w http.ResponseWriter, r *http.Request) {
			dur, err := time.ParseDuration(r.PathValue("duration"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			runShutdown(dur)
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
