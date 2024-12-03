
- First Downloads all titles
```
sh page-title-downloader.sh --lang=bn
```

- Go installation: https://go.dev/doc/install

- Second content download from titles
```
go run wiki-page-content-download.go --lang=bn  --input=./title-db/bn-titles.txt --output=./content-db/contents.txt
```
