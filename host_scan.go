package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/gookit/color"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type info struct {
	ip string
	host string
	url string
	status int
	length int64
	title string
	location string
	//Type string
}

func terminalOutput(i info){
	cyan := color.FgCyan.Render
	green := color.FgGreen.Render
	yellow := color.FgYellow.Render
	red := color.FgRed.Render
	//color.FgDefault.Render()
	var status string = strconv.Itoa(i.status)
	var title string = i.title
	var location = i.location
	var length = strconv.Itoa(int(i.length))
	if status[0] == '2' {
		status = green(status)
		title = green(title)
		length = green(length)
	}else if status[0] == '3' {
		status = yellow(status)
		title = yellow(title)
		length = yellow(length)
		location = yellow(location)
	}else{
		status = red(status)
		title = red(title)
		length = red(length)
	}

	if i.status/100 == 3{
		fmt.Printf("%s: %s\t%s: %s\t%s: %s\t%s: %s\t%s: %s\t%s: %s\n",
			cyan("IP"),i.ip,
			cyan("URL"),i.url,
			cyan("Status"),status,
			cyan("Length"),length,
			cyan("Title"),title,
			//cyan("Type"),i.Type,
			cyan("Location"),location,
		)
	}else{
		fmt.Printf("%s: %s\t%s: %s\t%s: %s\t%s: %s\t%s: %s\n",
			cyan("IP"),i.ip,
			cyan("URL"),i.url,
			cyan("Status"),status,
			cyan("Length"),length,
			cyan("Title"),title,
			//cyan("Type"),i.Type,
		)
	}
}


func write2File(w *bufio.Writer,i info)  {
	//for _,i:=range iList{
		var line string
		if i.status/100 == 3 {
			line = fmt.Sprintf("%s: %s\t%s: %s\t%s: %d\t%s: %d\t%s: %s\t%s: %s\n",
				"IP",i.ip,
				"URL",i.url,
				"Status",i.status,
				"Length",i.length,
				"Title",i.title,
				//cyan("Type"),i.Type,
				"Location",i.location,
			)
		}else{
			line = fmt.Sprintf("%s: %s\t%s: %s\t%s: %d\t%s: %d\t%s: %s\n",
				"IP",i.ip,
				"URL",i.url,
				"Status",i.status,
				"Length",i.length,
				"Title",i.title,
				//cyan("Type"),i.Type,
			)
		}
		_, err := w.WriteString(line)
		if err!=nil{
			panic(err)
		}
		w.Flush()
	//}
}


// get title
func getTitle(resp *http.Response)  (title string){
	respByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	reg, _ := regexp.Compile(`(?Ui:<title>[\s ]*([\s\S]*)[\s ]*</?title>)`)
	m := reg.FindStringSubmatch(string(respByte))
	if len(m) != 0 {
		title = strings.Replace(m[1], "\n", "", -1)
	}
	return title
}

func FileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}


var userAgent *string
var filter_length *string
func main() {
	// flag logic
	var timeout int
	ipFile:=flag.String("i","","IP list file (required)")
	hostFile:=flag.String("d","","Host/Domain list file (required)")
	filter_length=flag.String("fl","","Filter by length of response. Comma-separated list of length like 0,123")
	outputFile:=flag.String("output","","Output file")
	suffix:=flag.String("suffix","","Append a suffix to each line of the host list")
	threads:=flag.Int("threads",50,"Threads/Goroutine number")
	flag.IntVar(&timeout,"timeout",8,"Request timeout")
	redirect:=flag.Bool("redirect",false,"Follow redirects")
	userAgent=flag.String("ua","Mozilla/5.0(Linux;U;Android2.3.6;zh-cn;GT-S5660Build/GINGERBREAD)AppleWebKit/533.1(KHTML,likeGecko)Version/4.0MobileSafari/533.1MicroMessenger/4.5.255","User-Agent string")
	flag.Parse()
	if *ipFile=="" || *hostFile==""{
		fmt.Println("Use -h show help!")
		os.Exit(0)
	}
	// out to file pointer
	var w *bufio.Writer
	if *outputFile != ""{
		if FileExist(*outputFile){
			fmt.Println(*outputFile,"already exist!")
			os.Exit(-1)
		}
		f, err := os.OpenFile(*outputFile, os.O_CREATE|os.O_WRONLY, 0664)
		if err!=nil{
			panic(err)
		}
		defer f.Close()
		w = bufio.NewWriter(f)
	}

	// skip verify cert
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client:=&http.Client{
		Timeout: time.Duration(time.Duration(timeout) * time.Second),
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse /* 不进入重定向 */
		},
	}
	// follow redirect
	// fmt.Println(*redirect)
	if *redirect{
		client=&http.Client{
			Timeout: time.Duration(time.Duration(timeout) * time.Second),
			Transport: tr,
		}
	}

	// general brute list
	ipList:=file2List(*ipFile)
	hostList:=file2List(*hostFile)
	var bruteList [][]string
	for _,host:=range hostList{
			for _,ip:=range ipList{
				// fmt.Println(strings.TrimSpace(ip), strings.TrimSpace(host))
				bruteList = append(bruteList, []string{strings.TrimSpace(ip), strings.TrimSpace(host)})
			}
	}

	// start goroutine
	wg:=sync.WaitGroup{}
	limit:=make(chan bool,*threads) // workers count
	for _,item:=range bruteList{
		// fmt.Println(item)
		wg.Add(1)
		limit <- true
		ip,host:=item[0],item[1]
		go func(string,string, *bufio.Writer) {
			defer func() {
				wg.Done()
				<- limit
			}()
			var infoList []info
			if suf:= *suffix;suf != ""{
				if !strings.HasPrefix(*suffix,"."){
					suf = "." + *suffix
				}
				host += suf
			}
			infoList=sendRequests(client,ip,host)
			for _,i:=range infoList{
				if fl := *filter_length; fl != ""{
					if IsContain(strings.Split(fl,","),i.length){
						continue
					}
				}
				terminalOutput(i)
				// write to file
				if w!=nil{
					write2File(w,i)
				}
			}
		}(ip,host,w)
	}
	wg.Wait()
}


func file2List(fileName string) (text []string){
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	//content, err := ioutil.ReadAll(file)
	//if err !=nil{
	//	panic(err)
	//}
	//text = strings.Split(string(content),"\n")
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var line string =strings.TrimSpace(scanner.Text())
		if line!="" && !strings.HasPrefix(line,"//") && !strings.HasPrefix(line,"#"){
			text = append(text, line)
		}
	}
	return text
}


func sendRequests(client *http.Client,ip string, host string) (ret []info){
	schemaHttp := "http://"
	schemaHttps := "https://"
	for _,schema:=range []string{schemaHttp, schemaHttps}{
		req,_ := http.NewRequest(http.MethodGet,schema+ip+"/",nil)
		req.Host = host
		req.Header.Set("User-Agent",*userAgent)
		resp,err:=client.Do(req)
		if err!=nil{
			// log.Println(err) // cancel this comment show more info
			continue
		}
		var location string= ""
		if resp.StatusCode/100 == 3 {
			location = resp.Header.Get("Location")
		}
		ret = append(ret,info{
			ip:ip,
			host:host,
			url: schema+host+"/",
			status: resp.StatusCode,
			length: resp.ContentLength,
			title: getTitle(resp),
			location: location,
			//Type: resp.Header.Get("Content-Type"),
		})
	}
	return  ret
}

func IsContain(items []string, item int64) bool {
	for _, eachItem := range items {
		if eachItem == ""{
			continue
		}
		fli, err := strconv.ParseInt(eachItem,10,64)
		if err!=nil{
			fmt.Println("fl needs comma-separated list of length(number)")
			os.Exit(-1)
		}
		if fli == item {
			return true
		}
	}
	return false
}