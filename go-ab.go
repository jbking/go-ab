package main

import (
    "bufio"
    "flag"
    "fmt"
    "time"
    "log"
    "net"
    "net/url"
    "math"
)

var requests int
var concurrency int

func init() {
    flag.IntVar(&requests, "n", 1, "Number of requests to perform")
    flag.IntVar(&concurrency, "c", 1, "Number of multiple requests to make")
}

type ResponseResult struct {
    status string
    sec int64
    nsec int64
}

func get(address string, url *url.URL) (string, error) {
    conn, err := net.Dial("tcp", address)
    if err != nil {
        return "", err
    }
    defer conn.Close()

    fmt.Fprintf(conn, "GET %v HTTP/1.1\r\nHost: %v\r\n\r\n", url.Path, url.Host)
    status, err := bufio.NewReader(conn).ReadString('\n')
    if err != nil {
        return "", err
    }
    return status, nil
}

func run(c chan *ResponseResult, address string, url *url.URL, requests int) {
    for i := 0; i < requests; i++ {
        start := time.Now()
        status, err := get(address, url)
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

        rr := &ResponseResult{status, sec, nsec}
        c <- rr
    }
}

func main() {
    flag.Parse()

    n := requests * concurrency
    ul, err := url.Parse(flag.Arg(0))
    if err != nil {
        log.Fatal(err)
    }
    host, port, err := net.SplitHostPort(ul.Host)
    if err != nil {
        host = ul.Host
        port = "80"
    }

    addrs, err := net.LookupHost(host)
    if err != nil {
        log.Fatal(err)
    }

    address := fmt.Sprintf("%v:%v", addrs[0], port)
    c := make(chan *ResponseResult, n)
    for i := 0; i < concurrency; i++ {
        go run(c, address, ul, requests)
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
