package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type ResponseBlock struct {
	RequestTime  int64
	ResponseTime int64
	StatusCode   int
	URL          string
}

func SendRequests(url string, params map[string]string, SLO float64) int {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(SLO)*time.Second)
	defer cancel()

	// configure the request
	url_values := req.URL.Query()
	for k, v := range params {
		url_values.Add(k, v)
	}

	req.URL.RawQuery = url_values.Encode()

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return 500
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func DelayMicroseconds(us int64) {
	var tv syscall.Timeval
	_ = syscall.Gettimeofday(&tv)

	stratTick := int64(tv.Sec)*int64(1000000) + int64(tv.Usec) + us
	endTick := int64(0)
	for endTick < stratTick {
		_ = syscall.Gettimeofday(&tv)
		endTick = int64(tv.Sec)*int64(1000000) + int64(tv.Usec)
	}
}

func RequestHooker(timeTable []float64, syncTime int64, url string, SLO float64) []ResponseBlock {
	var wg sync.WaitGroup
	delayChan := make(chan struct{})
	resp := make([]ResponseBlock, len(timeTable))

	// wait until the sync time
	for time.Now().UnixNano()/1e6 < syncTime {
		continue
	}

	start := time.Now().UnixNano()
	for i, t := range timeTable {
		wg.Add(1)
		hasBeenDelayed := (time.Now().UnixNano() - start) / 1e3
		timeToDelay := int64(t*1e6) - hasBeenDelayed
		if timeToDelay < 0 {
			timeToDelay = 0
		}
		go func(timeToDelay, t int64, i int) {
			defer wg.Done()
			DelayMicroseconds(timeToDelay)
			delayChan <- struct{}{}
			t1 := time.Now().UnixNano()
			statusCode := SendRequests(url, map[string]string{}, SLO)
			t2 := time.Now().UnixNano()
			// fmt.Println("timeToDelay:", t, "time:", t2-start/1e3)
			resp[i] = ResponseBlock{
				RequestTime:  t1,
				ResponseTime: t2,
				StatusCode:   statusCode,
				URL:          url,
			}
		}(timeToDelay, int64(t*1e6), i)
		<-delayChan
	}
	wg.Wait()
	return resp
}

func ToCsv(filename string, responseBlock []ResponseBlock) {
	newFile, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		newFile.Close()
	}()
	newFile.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(newFile)
	header := []string{"RequestTime", "ResponseTime", "StatusCode", "URL"}
	w.Write(header)
	for _, resp := range responseBlock {
		context := []string{
			strconv.FormatInt(resp.RequestTime, 10),
			strconv.FormatInt(resp.ResponseTime, 10),
			strconv.Itoa(resp.StatusCode),
			resp.URL,
		}
		w.Write(context)
	}
	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}
	w.Flush()
}

func ReadFloats(r io.Reader) ([]float64, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)
	var result []float64
	for scanner.Scan() {
		x, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			return result, err
		}
		result = append(result, x)
	}
	return result, scanner.Err()
}

func main() {
	var name string
	var distName string
	var distDir string
	var url string
	var SLO float64
	var syncTime int64
	var dest string

	// Bind the flag
	flag.StringVar(&distName, "dist", "wula", "the name of dist file")
	flag.StringVar(&name, "name", "wula", "the name of data file")
	flag.StringVar(&distDir, "dir", "./", "the target directory of data file")
	flag.StringVar(&url, "url", "http://127.0.0.1:8080/", "the target url")
	flag.Float64Var(&SLO, "SLO", 10, "Service Level Objective")
	flag.Int64Var(&syncTime, "synctime", 0, "the synchronization timestamp of wula (milliseconds)")
	flag.StringVar(&dest, "dest", "request.csv", "the destination of request.csv")

	// Parse the flag
	flag.Parse()

	// create all directories in the path if not exists
	if _, err := os.Stat(distDir); os.IsNotExist(err) {
		os.MkdirAll(distDir, os.ModePerm)
	}

	// read the dist file
	distFile := filepath.Join(distDir, distName+"-dist.txt")
	f, err := os.Open(distFile)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	dist, err := ReadFloats(f)
	if err != nil {
		fmt.Println(err)
	}
	resp := RequestHooker(dist, syncTime, url, SLO)
	ToCsv(dest, resp)
}
