// +build darwin

package main

import (
	"os/exec"

	"github.com/gen2brain/beeep"
	"github.com/gen2brain/dlgs"
)

var osWithGUI = true

func openBrowser(url string) error {
	var cmd string
	var args []string
	cmd = "open"
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

func initPlatformSpecific() {

}
