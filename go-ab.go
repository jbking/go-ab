package main

import (
    "flag"
    "fmt"
    "time"
    "log"
    "net/http"
    "math"
)

var requests int
var concurrency int

func init() {
    flag.IntVar(&requests, "n", 1, "Number of requests to perform")
    flag.IntVar(&concurrency, "c", 1, "Number of multiple requests to make")
}

type ResponseResult struct {
    status int
    sec int64
    nsec int64
}

func run(c chan *ResponseResult, url string, requests int) {
    for i := 0; i < requests; i++ {
        start := time.Now()
        res, err := http.Get(url)
        if err != nil {
            log.Fatal(err)
        }
        end := time.Now()

        sec := end.Unix() - start.Unix()
        nsec := end.UnixNano() - start.UnixNano()
        if nsec < 0 {
            sec = sec - 1
            nsec = nsec + 999999999
        }

        rr := &ResponseResult{res.StatusCode, sec, nsec}
        c <- rr
    }
}

func main() {
    flag.Parse()

    n := requests * concurrency
    url := flag.Arg(0)
    if url == "" {
        log.Fatal("Specify URL")
    }

    c := make(chan *ResponseResult, n)
    for i := 0; i < concurrency; i++ {
        go run(c, url, requests)
    }
    min := math.MaxFloat64
    max := float64(0)
    for i := 0; i < n; i++ {
        rr := <-c
        sec_float := float64(rr.sec) + float64(rr.nsec) / float64(1000000000)
        if sec_float > max {
            max = sec_float
        }
        if sec_float < min {
            min = sec_float
        }
    }
    fmt.Printf("min: %v\n", min)
    fmt.Printf("max: %v\n", max)
}
