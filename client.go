package teapot

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/tedsuo/rata"
)

type Client interface {
	CreateWorkstation(request WorkstationCreateRequest) error
	DeleteWorkstation(name string) error
	AttachWorkstation(name string) (*websocket.Conn, error)
}

type client struct {
	httpClient *http.Client
	reqGen     *rata.RequestGenerator
}

func NewClient(url string) Client {
	return &client{
		httpClient: &http.Client{},
		reqGen:     rata.NewRequestGenerator(url, Routes),
	}
}

func (c *client) CreateWorkstation(request WorkstationCreateRequest) error {
	return c.doRequest(CreateWorkstationRoute, nil, nil, request, nil)
}

func (c *client) DeleteWorkstation(name string) error {
	return c.doRequest(DeleteWorkstationRoute, rata.Params{"name": name}, nil, nil, nil)
}

func (c *client) AttachWorkstation(name string) (*websocket.Conn, error) {
	return c.wsRequest(AttachWorkstationRoute, rata.Params{"name": name}, nil, nil)
}

func (c *client) doRequest(requestName string, params rata.Params, queryParams url.Values, request, response interface{}) error {
	requestJson, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := c.reqGen.CreateRequest(requestName, params, bytes.NewReader(requestJson))
	if err != nil {
		return err
	}

	req.URL.RawQuery = queryParams.Encode()
	req.ContentLength = int64(len(requestJson))

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode > 299 {
		errResponse := Error{}
		json.NewDecoder(res.Body).Decode(&errResponse)
		return errResponse
	}

	if response != nil {
		return json.NewDecoder(res.Body).Decode(&response)
	} else {
		return nil
	}
}

func (c *client) wsRequest(requestName string, params rata.Params, queryParams url.Values, request interface{}) (*websocket.Conn, error) {
	requestJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := c.reqGen.CreateRequest(requestName, params, bytes.NewReader(requestJson))
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = queryParams.Encode()
	req.ContentLength = int64(len(requestJson))
	req.URL.Scheme = "ws"

	if req.URL.User != nil {
		req.Header.Add("Authorization", basicAuth(req.URL.User))
		req.URL.User = nil
	}
	req.Header.Add("Origin", req.URL.String())

	conn, err := dialEndpoint(req.URL, nil)
	if err != nil {
		return nil, err
	}

	ws, res, err := websocket.NewClient(conn, req.URL, req.Header, 1024, 1024)
	if res.StatusCode > 299 {
		errResponse := Error{}
		json.NewDecoder(res.Body).Decode(&errResponse)
		return nil, errResponse
	}
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func basicAuth(user *url.Userinfo) string {
	username := user.Username()
	password, _ := user.Password()
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

var canonicalPortMap = map[string]string{
	"ws":    "80",
	"wss":   "4443",
	"http":  "80",
	"https": "443",
}

func dialEndpoint(url *url.URL, tlsConfig *tls.Config) (net.Conn, error) {
	addr := canonicalAddr(url)

	if url.Scheme == "https" {
		return tls.Dial("tcp", addr, tlsConfig)
	} else {
		return net.Dial("tcp", addr)
	}
}

func canonicalAddr(url *url.URL) string {
	host, port, err := net.SplitHostPort(url.Host)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			host = url.Host
			port = canonicalPortMap[url.Scheme]
		} else {
			log.Fatalln("invalid host:", err)
		}
	}

	return net.JoinHostPort(host, port)
}
