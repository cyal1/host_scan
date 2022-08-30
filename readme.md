## 这是一个用于IP和域名碰撞匹配访问的小工具，旨意用来匹配出渗透过程中需要绑定hosts才能访问的弱主机或内部系统。

## install
```bash
git clone https://github.com/cyal1/host_scan.git

cd host_scan

go build host_scan.go

go install .
```
## usage

`hostscan -i ip.txt -d host.txt`

It will be skipped If the line starts with `#` or `//` in ip.txt/host.txt

Before use host_can, recommend `sort -u ip.txt -o ip_uniq.txt`

```
Usage of hostscan:
  -d string
        Host/Domain list file (required)
  -fc string
        Filter by comma separated of status code (example '403,404')
  -fl string
        Filter by comma separated of content-length (example '133,14213')
  -fs string
        Filter by string in response, support for regex (example '(?i)(nginx|table)')
  -i string
        IP list file (required)
  -output string
        Output file
  -paths string
        Comma separated paths (example '/api/v1,/api/v2') (default "/")
  -redirect
        Follow redirects
  -suffix string
        Append a suffix to each line of the host list
  -threads int
        Threads/Goroutine number (default 50)
  -timeout int
        Request timeout (default 8)
  -ua string
        User-Agent string (default "Mozilla/5.0(Linux;U;Android2.3.6;zh-cn;GT-S5660Build/GINGERBREAD)AppleWebKit/533.1(KHTML,likeGecko)Version/4.0MobileSafari/533.1MicroMessenger/4.5.255")
```

![效果图](https://raw.githubusercontent.com/cyal1/host_scan/master/test.jpg)

## TODO

CIDR support
结果去重

## Reference
https://github.com/fofapro/Hosts_scan

