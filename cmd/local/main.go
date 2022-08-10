package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/mqyang56/gotunnel"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(flag.CommandLine)
	defer klog.Flush()
	flag.Parse()

	go cloud()
	edge()
}

func edge() {
	go func() {
		rp := httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   "10.11.0.213:9090",
		})
		http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			rp.ServeHTTP(writer, request)
		})
		err := http.ListenAndServe(":9091", nil)
		if err != nil {
			panic(err)
		}
	}()

	stopCh := make(chan struct{})
	defer close(stopCh)
	gotunnel.NewTunnel("127.0.0.1:9099", "abc", stopCh)
}

func cloud() {
	fs, err := gotunnel.NewFrontServer(":9099")
	if err != nil {
		panic(err)
	}

	mu := http.NewServeMux()
	mu.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		host := strings.Split(strings.Split(request.Host, ":")[0], ".")[0]
		req, err := http.NewRequest(request.Method, fmt.Sprintf("http://%s:`9091`%s", host, request.URL.String()), request.Body)
		req.Header = request.Header
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := fs.DoRequest(host, "9091", req)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		h := writer.Header()
		for k, v := range resp.Header {
			h[k] = v
		}

		data, _ := ioutil.ReadAll(resp.Body)
		_, err = writer.Write(data)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err = http.ListenAndServe(":8080", mu)
	if err != nil {
		panic(err)
	}
}
