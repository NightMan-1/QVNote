![Travis (.com)](https://img.shields.io/travis/com/NightMan-1/QVNote?style=flat-square) ![GitHub issues](https://img.shields.io/github/issues/NightMan-1/QVNote?style=flat-square) ![GitHub release](https://img.shields.io/github/release-pre/NightMan-1/QVNote?style=flat-square) ![GitHub (Pre-)Release Date](https://img.shields.io/github/release-date-pre/NightMan-1/QVNote?style=flat-square) ![GitHub All Releases](https://img.shields.io/github/downloads/NightMan-1/QVNote/total?style=flat-square)

# QVNote

The program for storing notes, pages of sites, personal knowledge base and any other text data

English/Russian languages

All data stored in JSON format (format based on [Quiver](http://happenapps.com/))

Written with GoLang and VueJS

**Requires Chrome/Chromium >= 70 to be installed**

More info here - https://qvnote.fsky.info/

![Screenshot eng](screenshot_eng.png)

## Usage

You can always download latest stable binary from here - https://github.com/NightMan-1/QVNote/releases/latest - or build from sources

## Build from source

#### Project request

GoLang >1.13 
NodeJS  >12.x

#### Project setup

```
git clone https://github.com/NightMan-1/QVNote
cd QVNote
go get -u github.com/go-bindata/go-bindata/...
go get -u github.com/blevesearch/bleve
go get -u github.com/blevesearch/snowballstem
go get -u github.com/dustin/go-humanize
go get -u github.com/imroc/req
go get -u github.com/json-iterator/go
go get -u github.com/kataras/iris
go get -u github.com/iris-contrib/middleware/cors
go get -u github.com/google/uuid
go get -u github.com/siddontang/ledisdb/config
go get -u github.com/siddontang/ledisdb/ledis
go get -u github.com/json-iterator/go
go get -u github.com/marcsauter/single
go get -u github.com/josephspurrier/goversioninfo/cmd/goversioninfo
go get -u github.com/syndtr/goleveldb/leveldb
go get -u github.com/go-ini/ini
go get -u github.com/zserge/lorca

npm install
```

Addition for Windows system:
```
go get -u github.com/gen2brain/beeep
go get -u github.com/gen2brain/dlgs
go get -u github.com/getlantern/systray
go get -u github.com/gonutz/w32
```

Addition for MacOS:
```
go get -u github.com/gen2brain/beeep
go get -u github.com/gen2brain/dlgs
go get -u github.com/getlantern/systray
```


#### Compiles for production
```
npm run build
go-bindata templates/... icon.ico
goversioninfo
go build
```
now you can run QVNote binary

#### GUI development

run server (QVNote.exe)
npm run serve
open http://localhost:8080

## Command line parameters:
    --help
        usage info
    --port=8000
        listen port
    --portable
        portable mode for Windows OS, data will be stored in app folder
    --server
        server mode without systray and other GUI
    --datadir
        data folder, default $HOME/.config/QVNote or %USERPROFILE%/.config/QVNote

Also you can you optional configuration file "config.ini"

