package pkg

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/go-resty/resty/v2"
	nurl "net/url"
	"reflect"
	"time"
)

// HTTP Simplifies making requests to AEM instance. Handles authentication and URI auto-escaping.
type HTTP struct {
	instance *Instance
	baseURL  string
}

func NewHTTP(instance *Instance, baseURL string) *HTTP {
	return &HTTP{instance: instance, baseURL: baseURL}
}

func (h *HTTP) Client() *resty.Client {
	cv := h.instance.manager.aem.config.Values()
	client := resty.New()
	client.SetBaseURL(h.baseURL)
	client.SetBasicAuth(h.instance.User(), h.instance.Password())
	client.SetDoNotParseResponse(true)
	client.SetTimeout(cv.GetDuration("instance.http.timeout"))
	client.SetDebug(cv.GetBool("instance.http.debug"))
	client.SetDisableWarn(cv.GetBool("instance.http.disable_warn"))
	if cv.GetBool("instance.http.ignore_ssl_errors") {
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}
	return client
}

func (h *HTTP) Request() *resty.Request {
	return h.Client().R().SetBasicAuth(h.instance.User(), h.instance.Password())
}

func (h *HTTP) RequestWithTimeout(timeout time.Duration) *resty.Request {
	client := h.Client()
	client.SetTimeout(timeout)
	return client.R()
}

func (h *HTTP) RequestFormData(props map[string]any) *resty.Request {
	request := h.Request()
	for k, v := range props {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			for i := 0; i < rv.Len(); i++ {
				request.FormData.Add(k, fmt.Sprintf("%v", rv.Index(i).Interface()))
			}
		} else {
			request.FormData.Add(k, fmt.Sprintf("%v", v))
		}
	}
	return request
}

func (h *HTTP) BasicAuthCredentials() string {
	auth := h.instance.user + ":" + h.instance.password

	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (h *HTTP) Port() string {
	urlConfig, _ := nurl.Parse(h.baseURL)
	port := urlConfig.Port()
	if port == "" {
		if urlConfig.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return port
}

func (h *HTTP) Hostname() string {
	urlConfig, _ := nurl.Parse(h.baseURL)
	return urlConfig.Hostname()
}

func (h *HTTP) BaseURL() string {
	return h.baseURL
}
