package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/lfxnxf/emo-frame/logging"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	defaultTimeout = 5
)

type client struct {
	ctx             context.Context
	url             string
	header          http.Header
	body            io.Reader
	method          string
	timeout         int64
	err             error
	respBody        *http.Response
	tlsClientConfig *tls.Config
}

func NewReq(ctx context.Context) *client {
	return &client{
		ctx:     ctx,
		timeout: defaultTimeout,
	}
}

func (c *client) Get(url string) *client {
	c.url = url
	c.method = http.MethodGet
	return c
}

func (c *client) Post(url string) *client {
	c.url = url
	c.method = http.MethodPost
	return c
}

func (c *client) WithHeader(k string, v interface{}) *client {
	if c.header == nil {
		c.header = http.Header{}
	}
	c.header.Add(k, fmt.Sprint(v))
	return c
}

func (c *client) WithHeaderMap(header map[string]interface{}) *client {
	for k, v := range header {
		c.header.Add(k, fmt.Sprint(v))
	}
	return c
}

func (c *client) WithHeaders(keyAndValues ...interface{}) *client {
	l := len(keyAndValues) - 1
	for i := 0; i < l; i += 2 {
		k := fmt.Sprint(keyAndValues[i])
		c.header.Add(k, fmt.Sprint(keyAndValues[i+1]))
	}
	if (l+1)%2 == 1 {
		logging.For(c.ctx, zap.String("func", "client.NewReq().XXX().WithHeaders")).Warnw("the keys are not aligned")
		k := fmt.Sprint(keyAndValues[l])
		c.header.Add(k, "")
	}
	return c
}

func (c *client) WithTimeout(timeout int64) *client {
	c.timeout = timeout
	return c
}

func (c *client) WithBody(body interface{}) *client {
	switch v := body.(type) {
	case io.Reader:
		buf, err := ioutil.ReadAll(v)
		if err != nil {
			c.err = err
			return c
		}
		c.body = bytes.NewReader(buf)
	case []byte:
		c.body = bytes.NewReader(v)
	case string:
		c.body = strings.NewReader(v)
	default:
		buf, err := jsoniter.Marshal(body)
		if err != nil {
			c.err = err
			return c
		}
		c.body = bytes.NewReader(buf)
	}
	return c
}

type option func(c *client)

func (c *client) Response() *client {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: c.tlsClientConfig,
			DialContext: (&net.Dialer{
				Timeout:   time.Duration(c.timeout) * time.Second,
				KeepAlive: time.Second * 5,
			}).DialContext,
			IdleConnTimeout:     time.Second * 5,
			MaxIdleConnsPerHost: 10,
		},
		Timeout: time.Duration(c.timeout) * time.Second,
	}

	if c.method == http.MethodGet {
		c.body = nil
	}
	req, err := http.NewRequest(c.method, c.url, c.body)
	if err != nil {
		c.err = err
		return c
	}
	req.Header = c.header
	resp, err := client.Do(req)
	if err != nil {
		c.err = err
		return c
	}
	c.respBody = resp
	return c
}

func (c *client) TLSClientConfig(conf *tls.Config) *client {
	c.tlsClientConfig = conf
	return c
}

func (c *client) ParseJson(data interface{}) error {
	return c.ParseDataJson(data)
}

func (c *client) ParseEmpty() error {
	return c.ParseDataJson(nil)
}

func (c *client) ParseDataJson(data interface{}) error {
	if c.err != nil {
		return c.err
	}
	defer func() {
		_ = c.respBody.Body.Close()
	}()

	if c.respBody.StatusCode != http.StatusOK {
		return errors.New(c.respBody.Status)
	}

	// 空解析
	if data == nil {
		return nil
	}

	body, err := ioutil.ReadAll(c.respBody.Body)
	if err != nil {
		return err
	}
	return jsoniter.Unmarshal(body, data)
}

func (c *client) ParseString(str *string) error {
	if c.err != nil {
		return c.err
	}
	defer func() {
		_ = c.respBody.Body.Close()
	}()

	if c.respBody.StatusCode != http.StatusOK {
		return errors.New(c.respBody.Status)
	}

	body, err := ioutil.ReadAll(c.respBody.Body)
	if err != nil {
		return err
	}

	*str = string(body)
	return nil
}
