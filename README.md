- Go installation: https://go.dev/doc/install

```
wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz
# Add in ~/.bashrc
export PATH=$PATH:/usr/local/go/bin
source ~/.bashrc
```

- Downloads all titles
```
sh page-title-downloader.sh --lang=bn
```

- Download page coontent from title
```
nohup go run wiki-page-content-download.go --input=./inputs/titles-part-2.txt --output=./outputs/content-2.txt --username=xxx --password=xxx > output.log 2>&1 &
```


- top word find
```
grep -o -P '[\x{0980}-\x{09FF}]+' merged.txt | sort | uniq -c | sort -nr | head -n 10
```