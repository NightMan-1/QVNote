// +build linux

package main

func openBrowser(url string) error {
	return nil
}

func showNotification(messageText string, messageType string) {
	//do nothing
}
func runSystray() {
}

func initPlatformSpecific() {
	configGlobal.atStartOpenBrowser = false

}
