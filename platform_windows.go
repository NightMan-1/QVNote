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

//check console and start new one if not present
func initConsole() {
	modkernel32 := syscall.NewLazyDLL("kernel32.dll")
	procAllocConsole := modkernel32.NewProc("AllocConsole")
	r0, _, _ := syscall.Syscall(procAllocConsole.Addr(), 0, 0, 0, 0)
	if r0 == 0 { // Allocation failed, probably process already has a console
		//fmt.Printf("Could not allocate console: %s. Check build flags..", err0)
		configGlobal.consoleControl = false
	} else {
		hout, err1 := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
		hin, err2 := syscall.GetStdHandle(syscall.STD_INPUT_HANDLE)
		if err1 == nil && err2 == nil { // nowhere to print the error
			os.Stdout = os.NewFile(uintptr(hout), "/dev/stdout")
			os.Stdin = os.NewFile(uintptr(hin), "/dev/stdin")
			configGlobal.consoleControl = true

			// needed for show/hide console
			getConsoleWindow := modkernel32.NewProc("GetConsoleWindow")
			if getConsoleWindow.Find() == nil {
				showWindow = syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow")
				if showWindow.Find() == nil {
					hwnd, _, _ = getConsoleWindow.Call()
				}
			}

		}
	}
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
