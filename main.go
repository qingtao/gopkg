package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

var (
	// git addr, http[s]://example.org
	backend  = flag.String("backend", "", "git's address")
	frontend = flag.String("frontend", "", "local addr")
	addr     = flag.String("addr", "127.0.0.1:80", "listen on")
	debug    = flag.Bool("debug", false, "debug")
)

type gitReverseServer struct {
	backend  *url.URL
	frontend *url.URL
	rs       *httputil.ReverseProxy
}

func debugf(formart string, v ...any) {
	if !*debug {
		return
	}
	log.Printf(formart, v...)
}

func (a *gitReverseServer) modifyResponse() func(r *http.Response) error {
	return func(r *http.Response) error {
		if r.Request.FormValue("go-get") != "1" {
			return nil
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			r.Body.Close()
			return err
		}
		r.Body.Close()
		body = bytes.ReplaceAll(body, []byte(a.backend.Host), []byte(a.frontend.Host))
		if a.backend.Scheme != a.frontend.Scheme {
			body = bytes.ReplaceAll(body, []byte(a.backend.Scheme), []byte(a.frontend.Scheme))
		}
		r.ContentLength = int64(len(body))
		r.Header.Set("Content-Length", strconv.FormatInt(r.ContentLength, 10))
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		return nil
	}

}

func createGitReverseServer(backend, frontend string) (*gitReverseServer, error) {
	b, err := url.Parse(backend)
	if err != nil {
		return nil, fmt.Errorf("parse backend url %w", err)
	}
	// if b.Scheme == "https" && b.Port() == "" {
	// 	b.Host += ":443"
	// }
	f, err := url.Parse(frontend)
	if err != nil {
		return nil, fmt.Errorf("parse frontend url %w", err)
	}
	s := &gitReverseServer{
		backend:  b,
		frontend: f,
	}
	rs := &httputil.ReverseProxy{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(s.backend)
			pr.Out.Host = pr.In.Host
			debugf("pr.in = %s, pr.out = %s", pr.In.URL, pr.Out.URL)
		},
		ModifyResponse: s.modifyResponse(),
	}
	s.rs = rs

	return s, nil
}

func (a *gitReverseServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.rs.ServeHTTP(w, r)
}

func init() {
	flag.Parse()
	log.SetPrefix("gopkg ")
}

func main() {
	s, err := createGitReverseServer(*backend, *frontend)
	if err != nil {
		log.Fatalln(err)
	}
	if err := http.ListenAndServe(*addr, s); err != nil {
		log.Fatalln(err)
	}
}
