// +build windows

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/gen2brain/dlgs"
	"github.com/getlantern/systray"
	"github.com/gonutz/w32"
	"github.com/zserge/lorca"
)

var consoleWindows w32.HWND
var consoleWindowsVisible bool
var consolePresent bool

func openBrowser(url string) error {
	var cmd string
	var args []string

	cmd = "cmd"
	args = []string{"/c", "start"}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func showNotification(messageText string, messageType string) {
	if configGlobal.cmdServerMode {
		return
	}
	switch messageType {
	case "dialog_warning":
		dlgs.Warning("QVNote error!", messageText)
	case "notify":
		if runtime.GOOS == string("windows") {
			var tmpIcon *os.File
			iconData, _ := Asset("icon.ico")
			tmpIcon, _ = ioutil.TempFile("", "icon.*.ico")
			tmpIcon.Write(iconData)
			tmpIcon.Close()
			beeep.Notify("QVNote", messageText, tmpIcon.Name())
			time.Sleep(50 * time.Millisecond)
			os.Remove(tmpIcon.Name())
		} else {
			beeep.Notify("QVNote", messageText, "") // icon not work on MacOS
		}

	}
}

func consoleShow() {
	if consoleWindows != 0 {
		_, consoleProcID := w32.GetWindowThreadProcessId(consoleWindows)
		if w32.GetCurrentProcessId() == consoleProcID {
			w32.ShowWindowAsync(consoleWindows, w32.SW_SHOW)
			consoleWindowsVisible = true
		}
	}

}

func consoleHide() {
	if consoleWindows != 0 {
		_, consoleProcID := w32.GetWindowThreadProcessId(consoleWindows)
		if w32.GetCurrentProcessId() == consoleProcID {
			w32.ShowWindowAsync(consoleWindows, w32.SW_HIDE)
			consoleWindowsVisible = false
		}
	}
}

func onReadySysTray() {
	iconData, _ := Asset("icon.ico")

	systray.SetIcon(iconData)
	systray.SetTitle("QVNote")

	mBrowser := systray.AddMenuItem("Open browser", "open default browser with this app page")
	mOpenLorcaGUI := systray.AddMenuItem("Run independent mode", "Open Chrome based GUI")
	mRelod := systray.AddMenuItem("Reload notes", "may be slow")
	mShowConsoleHide := systray.AddMenuItem("Hide console", "debug")
	mShowConsoleShow := systray.AddMenuItem("Show console", "debug")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	if configGlobal.consolePresent {
		if consoleWindowsVisible {
			mShowConsoleShow.Hide()
		} else {
			mShowConsoleHide.Hide()
		}
	} else {
		mShowConsoleShow.Hide()
		mShowConsoleHide.Hide()
	}

	for {
		select {
		case <-mShowConsoleHide.ClickedCh:
			consoleHide()
			mShowConsoleHide.Hide()
			mShowConsoleShow.Show()
		case <-mShowConsoleShow.ClickedCh:
			consoleShow()
			mShowConsoleShow.Hide()
			mShowConsoleHide.Show()
		case <-mBrowser.ClickedCh:
			openBrowser("http://localhost:" + configGlobal.cmdPort + "/")
		case <-mRelod.ClickedCh:
			FindAllNotes()
			beeep.Notify("QVNote", "All data reloaded", "")
		case <-mQuit.ClickedCh:
			systray.Quit()
			beeep.Notify("QVNote", "Good buy!", "")
			fmt.Println("Good buy!")
		case <-mOpenLorcaGUI.ClickedCh:
			startStadaloneGUI()
		}
	}

}

func onExitSysTray() {
	// clean up here
	os.Exit(0)
}

func runSystray() {
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

func initPlatformSpecific() {
	if configGlobal.cmdServerMode {
		return
	}

	consoleWindows = w32.GetConsoleWindow()
	consoleWindowsVisible = true

	_, consoleProcID := w32.GetWindowThreadProcessId(consoleWindows)
	if w32.GetCurrentProcessId() == consoleProcID {
		configGlobal.consolePresent = true
	} else {
		configGlobal.consolePresent = false
	}

	if configGlobal.appStartingMode == "independent" {
		configGlobal.atStartShowConsole = false
	}

	if configGlobal.atStartShowConsole == false {
		consoleHide()
	}

	if configGlobal.appStartingMode != "independent" {
		go systray.Run(onReadySysTray, onExitSysTray)
	}

}
