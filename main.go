package main

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/aliyun/fc-runtime-go-sdk/fc"
)

func upstream(req *http.Request) *url.URL {
	// req.URL.Path client name
	url, _ := url.Parse(os.Getenv("UPSTREAM") + req.URL.Path)
	return url
}

func copy(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func proxy(req *http.Request, w http.ResponseWriter) error {
	client := &http.Client{}
	if resp, err := client.Do(req); err != nil {
		return err
	} else {
		defer resp.Body.Close()
		copy(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		return err
	}
}

func handleRequestGet(w http.ResponseWriter, req *http.Request) error {
	q := req.URL.Query()
	dns := q.Get("dns")

	if dns == "" {
		// w.Write([]byte("bad dns!"))
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	if req.Header.Get("accept") != "application/dns-message" {
		// w.Write([]byte("bad header: " + req.Header.Get("accept")))
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	// send
	url := upstream(req)
	request, _ := http.NewRequest("GET", url.String()+"?dns="+dns, nil)
	copy(request.Header, req.Header)
	request.Header.Add("host", url.Host)
	return proxy(request, w)
}

func handleRequestPost(w http.ResponseWriter, req *http.Request) error {
	if req.Header.Get("content-type") != "application/dns-message" {
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	// send
	url := upstream(req)
	request, _ := http.NewRequest("POST", url.String(), req.Body)
	request.Header.Add("accept", "application/dns-message")
	req.Header.Add("content-type", "application/dns-message")
	request.Header.Add("host", url.Host)

	return proxy(request, w)
}

func HandleHttpRequest(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
	switch req.Method {
	case "GET":
		return handleRequestGet(w, req)
	case "POST":
		return handleRequestPost(w, req)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

func main() {
	fc.StartHttp(HandleHttpRequest)
}
