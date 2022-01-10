//go:build linux

package main

func openBrowser(url string) error {
	return nil
}

func showNotificationDialog(messageText string) {
	//do nothing
}
func runSystray() {
}

func initPlatformSpecific() {
	configGlobal.atStartOpenBrowser = false
}
