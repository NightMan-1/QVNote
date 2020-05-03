systray is a cross-platform Go library to place an icon and menu in the notification area.

## Features

* Supported on Windows, macOS, and Linux
* Menu items can be checked and/or disabled
* Methods may be called from any Goroutine

## API

```go
func main() {
	// Should be called at the very beginning of main().
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("Awesome App")
	systray.SetTooltip("Pretty awesome超级棒")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	// Sets the icon of a menu item. Only available on Mac.
	mQuit.SetIcon(icon.Data)
}

func onExit() {
	// clean up here
}
```

## Try the example app!

Have go v1.12+ or higher installed? Here's an example to get started on macOS:

```sh
git clone https://github.com/getlantern/systray
cd example
env GO111MODULE=on go build
./example
```

On Windows, you should build like this:

```
env GO111MODULE=on go build -ldflags "-H=windowsgui"
```

The following text will then appear on the console:


```sh
go: finding github.com/skratchdot/open-golang latest
go: finding github.com/getlantern/systray latest
go: finding github.com/getlantern/golog latest
```

Now look for *Awesome App* in your menu bar!

![Awesome App screenshot](example/screenshot.png)

## Platform notes

### Linux

* Building apps requires the `gtk3`, `libappindicator3` and `libwebkit2gtk-4.0-dev` development headers to be installed. For Debian or Ubuntu, you can may install these using:

```sh
sudo apt-get install libgtk-3-dev libappindicator3-dev libwebkit2gtk-4.0-dev
```

To build the example, run `go build example/main.go`

* Checked menu items are not yet implemented


### Windows

* To avoid opening a console at application startup, use these compile flags:

```sh
go build -ldflags -H=windowsgui
```

You'll also need to include a Manifest for your windows app. Assuming your executable is named `app.exe`, create a file `app.manifest` like this:

```
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0">
    <assemblyIdentity version="1.0.0.0" processorArchitecture="*" name="App Name Here" type="win32"/>
    <dependency>
        <dependentAssembly>
            <assemblyIdentity type="win32" name="Microsoft.Windows.Common-Controls" version="6.0.0.0" processorArchitecture="*" publicKeyToken="6595b64144ccf1df" language="*"/>
        </dependentAssembly>
    </dependency>
    <application xmlns="urn:schemas-microsoft-com:asm.v3">
        <windowsSettings>
            <dpiAwareness xmlns="http://schemas.microsoft.com/SMI/2016/WindowsSettings">PerMonitorV2, PerMonitor</dpiAwareness>
            <dpiAware xmlns="http://schemas.microsoft.com/SMI/2005/WindowsSettings">True</dpiAware>
        </windowsSettings>
    </application>
</assembly>
```

Then either compile the manifest using the [rsrc](https://github.com/akavel/rsrc) tool, like this:

```
go get github.com/akavel/rsrc
rsrc -manifest app.manifest -o rsrc.syso
```

or rename the app.manifest file to app.exe.manifest and distribute it with the application instead.

### macOS

On macOS, you will need to create an application bundle to wrap the binary; simply folders with the following minimal structure and assets:

```
SystrayApp.app/
  Contents/
    Info.plist
    MacOS/
      go-executable
    Resources/
      SystrayApp.icns
```

When running as an app bundle, you may want to add one or both of the following to your Info.plist:

```xml
<!-- avoid having a blurry icon and text -->
	<key>NSHighResolutionCapable</key>
	<string>True</string>

	<!-- avoid showing the app on the Dock -->
	<key>LSUIElement</key>
	<string>1</string>
```

Consult the [Official Apple Documentation here](https://developer.apple.com/library/archive/documentation/CoreFoundation/Conceptual/CFBundles/BundleTypes/BundleTypes.html#//apple_ref/doc/uid/10000123i-CH101-SW1).

## Credits

- https://github.com/xilp/systray
- https://github.com/cratonica/trayhost
