// +build windows

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/gen2brain/beeep"
	"github.com/gen2brain/dlgs"
	"github.com/getlantern/systray"
	"github.com/gonutz/w32"
)

var osWithGUI = true
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
	go func() {
		<-mBrowser.ClickedCh
		go openBrowser("http://localhost:8000/")
	}()

	if consolePresent {
		mShowConsoleHide := systray.AddMenuItem("Hide console", "debug")
		mShowConsoleShow := systray.AddMenuItem("Show console", "debug")
		go func() {
			<-mShowConsoleHide.ClickedCh
			consoleHide()
			mShowConsoleHide.Hide()
			mShowConsoleShow.Show()
		}()
		go func() {
			<-mShowConsoleShow.ClickedCh
			consoleShow()
			mShowConsoleShow.Hide()
			mShowConsoleHide.Show()
		}()
		if consoleWindowsVisible {
			mShowConsoleShow.Hide()
		} else {
			mShowConsoleHide.Hide()
		}
	}

	mRelod := systray.AddMenuItem("Reload notes", "may be slow")
	go func() {
		<-mRelod.ClickedCh
		FindAllNotes()
		beeep.Notify("QVNote", "All data reloaded", "")
	}()

	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
		beeep.Notify("QVNote", "Good buy!", "")
		fmt.Println("Good buy!")
	}()

}

func onExitSysTray() {
	// clean up here
	os.Exit(0)
}

func initPlatformSpecific() {
	if configGlobal.cmdServerMode {
		return
	}

	consoleWindows = w32.GetConsoleWindow()
	consoleWindowsVisible = true

	_, consoleProcID := w32.GetWindowThreadProcessId(consoleWindows)
	if w32.GetCurrentProcessId() == consoleProcID {
		consolePresent = true
	} else {
		consolePresent = false
	}

	if configGlobal.atStartShowConsole == false {
		consoleHide()
	}
	go systray.Run(onReadySysTray, onExitSysTray)

}
