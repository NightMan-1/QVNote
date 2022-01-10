//go:build darwin

package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/getlantern/systray"
)

func openBrowser(url string) error {
	return exec.Command("open", url).Start()
}

func onReadySysTray() {
	iconData, _ := Asset("../icon.ico")

	systray.SetIcon(iconData)
	//systray.SetTitle("QVNote")

	mBrowser := systray.AddMenuItem("Open browser", "open default browser with this app page")
	mRelod := systray.AddMenuItem("Reload notes", "may be slow")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	for {
		select {
		case <-mBrowser.ClickedCh:
			openBrowser("http://localhost:" + configGlobal.cmdPort + "/")
		case <-mRelod.ClickedCh:
			url := "http://localhost:" + configGlobal.cmdPort + "/api/refresh_data.json"
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
			fmt.Println("Good buy!")
			os.Exit(0)
		}
	}
}

func runSystray() {
	systray.Run(onReadySysTray, onExitSysTray)
}

func onExitSysTray() {
	// clean up here
	client := http.Client{}
	client.Get("http://localhost:" + configGlobal.cmdPort + "/api/exit")
	os.Exit(0)
}

func initConsole() {}

func initPlatformSpecific() error {
	if configGlobal.cmdServerMode {
		return nil
	}

	if configGlobal.appStartingMode != "independent" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		systrayProcess = exec.Command(os.Args[0], "--systray", configGlobal.cmdPort)
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
