// +build darwin

package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/gen2brain/beeep"
	"github.com/gen2brain/dlgs"
	"github.com/getlantern/systray"
)

var osWithGUI = true

func openBrowser(url string) error {
	return exec.Command("open", url).Start()
}

func showNotification(message_text string, message_type string) {
	if configGlobal.cmdServerMode {
		return
	}
	switch message_type {
	case "dialog_warning":
		dlgs.Warning("QVNote error!", message_text)
	case "notify":
		beeep.Notify("QVNote", message_text, "")

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
			openBrowser("http://localhost:8000/")
		case <-mRelod.ClickedCh:
			url := "http://localhost:8000/api/refresh_data.json"
			var jsonStr = []byte(`{"action":"reload"}`)
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
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

func initPlatformSpecific() error {
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
	return nil
}
