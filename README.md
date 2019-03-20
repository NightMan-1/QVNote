# QVNote

The program for storing notes, pages of sites, personal knowledge base.

English/Russian languages

All data stored in Quiver like format (http://happenapps.com/)

GoLang server + VueJS frontend

![Screenshot eng](screenshot_eng.png)

## Project request

GoLang >1.10  
NodeJS  >10.x

## Project setup

Download  
Unzip archive  
_(you also need install git command line program for download the source)_

```
go get -u github.com/jteeuwen/go-bindata/...
go get -u github.com/blevesearch/bleve
go get -u github.com/blevesearch/snowballstem
go get -u github.com/dustin/go-humanize
go get -u github.com/imroc/req
go get -u github.com/json-iterator/go
go get -u github.com/kataras/iris
go get -u github.com/iris-contrib/middleware/cors
go get -u github.com/gofrs/uuid
go get -u github.com/siddontang/ledisdb/config
go get -u github.com/siddontang/ledisdb/ledis
go get -u github.com/json-iterator/go
go get -u github.com/marcsauter/single
go get -u github.com/josephspurrier/goversioninfo/cmd/goversioninfo
go get -u github.com/syndtr/goleveldb/leveldb

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



## Compiles for production
```
npm run build
go-bindata templates/... icon.ico
goversioninfo
go build
```
now you can run QVNote binary and open http://localhost:8000 in your browser

## GUI development

run server (QVNote.exe)  
npm run serve  
open http://localhost:8080


## Known issues:
1) systray icon work only in Windows
