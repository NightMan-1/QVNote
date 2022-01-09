//go:build darwin
// +build darwin

package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"github.com/gen2brain/beeep"
	"github.com/gen2brain/dlgs"
	"github.com/getlantern/systray"
	"github.com/zserge/lorca"
)

func openBrowser(url string) error {
	return exec.Command("open", url).Start()
}

func showNotification(messageText string, messageType string) {
	if configGlobal.cmdServerMode {
		return
	}
	switch messageType {
	case "dialog_warning":
		dlgs.Warning("QVNote error!", messageText)
	case "notify":
		beeep.Notify("QVNote", messageText, "")

	}
}

func onReadySysTray() {
	iconData, _ := Asset("icon.ico")

	systray.SetIcon(iconData)
	//systray.SetTitle("QVNote")

	mBrowser := systray.AddMenuItem("Open browser", "open default browser with this app page")
	mRelod := systray.AddMenuItem("Reload notes", "may be slow")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	for {
		select {
		case <-mBrowser.ClickedCh:
			//TODO dinamic port
			openBrowser("http://localhost:8000/")
		case <-mRelod.ClickedCh:
			url := "http://localhost:8000/api/refresh_data.json"
			var jsonStr = []byte(`{"action":"reload"}`)
			req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			resp.Body.Close()
		case <-mQuit.ClickedCh:
			systray.Quit()
			beeep.Notify("QVNote", "Good buy!", "")
			fmt.Println("Good buy!")
		}
	}
}

func runSystray() {
	systray.Run(onReadySysTray, onExitSysTray)
}

func onExitSysTray() {
	// clean up here
	client := http.Client{}
	client.Get("http://localhost:8000/api/exit")
	os.Exit(0)
}

func startStadaloneGUI() {
	// Create UI with basic HTML passed via data URI
	ui, err := lorca.New("data:text/html,"+url.PathEscape(`<html><head><title>QVNote</title></head><body>Loading...</body></html>`), "", 1380, 768)
	if err != nil {
		showNotification("Can not start Google Chrome", "dialog_warning")
		log.Fatalf("Can not start Google Chrome: %v", err)
		os.Exit(1)
	}
	defer ui.Close()
	ui.Load("http://localhost:8000")
	// Wait until UI window is closed
	<-ui.Done()
}

func initPlatformSpecific() error {
	if configGlobal.appStartingMode != "independent" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		systrayProcess = exec.Command(os.Args[0], "--systray")
		systrayProcess.Dir = cwd
		err = systrayProcess.Start()
		if err != nil {
			return err
		}
		//systrayProcess.Process.Kill()
		//systrayProcess.Process.Release()
	}
	return nil
}
