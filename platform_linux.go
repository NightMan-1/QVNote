//go:build linux

package main

func openBrowser(url string) error {
	return nil
}

func runSystray() {}

func initConsole() {}

func initPlatformSpecific() {
	configGlobal.atStartOpenBrowser = false
}
