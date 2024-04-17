package service

import (
	"bytes"
	"context"
	"github.com/Rorical/MirrRo/sansor"
	"github.com/Rorical/MirrRo/service/utils"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

func NewProxy(parsedURL *url.URL, requestModifier func(req *http.Request) (bool, error), responseModifier func(resp *http.Response) error) (func(http.ResponseWriter, *http.Request), error) {
	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	proxy.ModifyResponse = responseModifier
	handler := func(w http.ResponseWriter, r *http.Request) {
		banned, err := requestModifier(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if banned {
			http.Error(w, "This page is banned.", http.StatusForbidden)
			return
		}
		proxy.ServeHTTP(w, r)
	}
	return handler, nil
}

func GetRealIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarder-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}

type Router struct {
	proxies map[string]func(http.ResponseWriter, *http.Request)
}

func NewRouter(config *Config) (*Router, error) {
	router := &Router{
		proxies: make(map[string]func(http.ResponseWriter, *http.Request)),
	}
	sansorClient, err := sansor.NewSansorClient(config.SansorURI)
	if err != nil {
		return nil, err
	}
	replaceURIDFA := utils.NewDFA()
	replaceURIDFA.Build(config.Redirects)
	responseModifier := func(resp *http.Response) error {
		resp.Header.Set("X-Proxy", "Rorical")
		if resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusTemporaryRedirect {
			location := resp.Header.Get("Location")
			locationURL, err := url.Parse(location)
			if err != nil {
				return err
			}
			if redirect, ok := config.Redirects[locationURL.Host]; ok {
				locationURL.Host = redirect
				resp.Header.Set("Location", locationURL.String())
			}
		}
		contentType := resp.Header.Get("content-type")
		var bodyContent []byte
		var err error
		if contentType == "" {
			bodyContent, err = io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			contentType = http.DetectContentType(bodyContent)
		}

		if strings.Contains(contentType, ";") {
			contentType = strings.Split(contentType, ";")[0]
		}

		switch contentType {
		case "text/html", "text/css", "application/javascript", "application/json", "application/xml", "application/xhtml+xml", "image/svg+xml":
			if len(bodyContent) == 0 {
				bodyContent, err = io.ReadAll(resp.Body)
				if err != nil {
					return err
				}
			}
			isBan, err := sansorClient.TextReview(context.TODO(), utils.Bytes2String(bodyContent))
			if err != nil {
				return err
			}
			if isBan {
				resp.StatusCode = http.StatusForbidden
				resp.Header.Set("Content-Type", "text/plain")
				bodyContent = []byte("This page is banned.")
			}
			bodyContent = utils.String2Bytes(replaceURIDFA.ReplaceAll(utils.Bytes2String(bodyContent)))
			resp.Header.Set("Content-Length", strconv.Itoa(len(bodyContent)))
			break
		default:
			break
		}

		if len(bodyContent) > 0 {
			resp.Body = io.NopCloser(bytes.NewReader(bodyContent))
		}
		return nil
	}
	for host, target := range config.Mirrors {
		parsedURL, err := url.Parse(target)
		if err != nil {
			return nil, err
		}
		requestModifier := func(req *http.Request) (bool, error) {
			realIP := GetRealIP(req)
			req.Header.Set("X-Forwarded-For", realIP)
			req.Header.Set("X-Real-IP", realIP)
			if req.URL.Path == "" {
				req.URL.Path = "/"
			}
			req.Host = parsedURL.Host
			review, err := sansorClient.TextReview(context.TODO(), req.URL.String())
			if err != nil {
				return false, err
			}
			return review, nil
		}
		proxy, err := NewProxy(parsedURL, requestModifier, responseModifier)
		if err != nil {
			log.Fatal(err)
		}
		router.proxies[host] = proxy
	}
	return router, nil
}

func (r *Router) Handle(w http.ResponseWriter, req *http.Request) {
	proxy, ok := r.proxies[req.Host]
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	proxy(w, req)
}

func Listen(configPath string) error {
	cfg, err := ReadConfig(configPath)
	if err != nil {
		return err
	}
	router, err := NewRouter(cfg)
	if err != nil {
		return err
	}
	http.HandleFunc("/", router.Handle)
	return http.ListenAndServe("127.0.0.1:8999", nil)
}
