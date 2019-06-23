// +build linux

package main

var osWithGUI = false

func openBrowser(url string) error {
	return nil
}

func showNotification(message_text string, message_type string) {
	//do nothing
}
func runSystray() {
}

func initPlatformSpecific() {
	configGlobal.atStartOpenBrowser = false

}
