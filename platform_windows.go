//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/getlantern/systray"
)

var consoleWindowsVisible bool

var showWindow *syscall.LazyProc
var hwnd uintptr

func openBrowser(url string) error {
	return exec.Command("cmd", "/c", "start", url).Start()
}

func consoleShow() {
	if hwnd == 0 {
		return
	}
	showWindow.Call(hwnd, 9)
	consoleWindowsVisible = true
}

func consoleHide() {
	if hwnd == 0 {
		return
	}
	showWindow.Call(hwnd, 0)
	consoleWindowsVisible = false
}

func onReadySysTray() {
	iconData, _ := Asset("../icon.ico")

	systray.SetIcon(iconData)
	systray.SetTitle("QVNote")

	mBrowser := systray.AddMenuItem("Open browser", "open default browser with this app page")
	mOpenLorcaGUI := systray.AddMenuItem("Run independent mode", "Open Chrome based GUI")

	mRelod := systray.AddMenuItem("Reload notes", "may be slow")
	mRelod.SetIcon(iconRepeat)

	mShowConsoleHide := systray.AddMenuItem("Hide console", "debug")
	mShowConsoleHide.SetIcon(iconWindowsX)
	mShowConsoleShow := systray.AddMenuItem("Show console", "debug")
	mShowConsoleShow.SetIcon(iconWindows)

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	mQuit.SetIcon(iconPower)

	if configGlobal.consoleControl {
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
		case <-mQuit.ClickedCh:
			fmt.Println("Good buy!")
			if configGlobal.consoleControl {
				os.Exit(0)
			} else {
				SendInterrupt()
			}
		case <-mOpenLorcaGUI.ClickedCh:
			systray.Quit()
			startStadaloneGUI()
		}
	}

}

//https://stackoverflow.com/questions/40498371/how-to-send-an-interrupt-signal
func SendInterrupt() error {
	d, e := syscall.LoadDLL("kernel32.dll")
	if e != nil {
		return fmt.Errorf("LoadDLL: %v", e)
	}
	p, e := d.FindProc("GenerateConsoleCtrlEvent")
	if e != nil {
		return fmt.Errorf("FindProc: %v", e)
	}
	r, _, e := p.Call(syscall.CTRL_BREAK_EVENT, uintptr(syscall.Getpid()))
	if r == 0 {
		return fmt.Errorf("GenerateConsoleCtrlEvent: %v", e)
	}
	return nil
}

func runSystray() {
}

func initPlatformSpecific() {
	if configGlobal.cmdServerMode {
		configGlobal.consoleControl = false
		return
	}

	if configGlobal.appStartingMode == "independent" {
		configGlobal.consoleControl = true
	} else {
		go systray.Run(onReadySysTray, nil)
	}

	if configGlobal.consoleControl {
		consoleHide()
	}

}
