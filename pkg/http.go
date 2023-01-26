package pkg

import (
	"encoding/base64"
	"fmt"
	"github.com/go-resty/resty/v2"
	nurl "net/url"
	"reflect"
)

// HTTP Simplifies making requests to AEM instance. Handles authentication and URI auto-escaping.
type HTTP struct {
	client   *resty.Client
	instance *Instance
}

func NewHTTP(instance *Instance, baseURL string) *HTTP {
	return &HTTP{
		client:   newInstanceHTTPClient(baseURL),
		instance: instance,
	}
}

func newInstanceHTTPClient(baseURL string) *resty.Client {
	client := resty.New()
	client.SetBaseURL(baseURL)
	client.SetDisableWarn(true)
	client.SetDoNotParseResponse(true)
	return client
}

func (hc *HTTP) Request() *resty.Request {
	return hc.client.R().SetBasicAuth(hc.instance.User(), hc.instance.Password())
}

func (hc *HTTP) RequestFormData(props map[string]any) *resty.Request {
	request := hc.Request()
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

func (hc *HTTP) BaseURL() string {
	return hc.client.BaseURL
}

func (hc *HTTP) BasicAuthCredentials() string {
	auth := hc.instance.user + ":" + hc.instance.password

	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (hc *HTTP) Port() string {
	urlConfig, _ := nurl.Parse(hc.client.BaseURL)
	return urlConfig.Port()
}

func (hc *HTTP) Hostname() string {
	urlConfig, _ := nurl.Parse(hc.client.BaseURL)
	return urlConfig.Hostname()
}
