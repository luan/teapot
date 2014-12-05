package teapot

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/tedsuo/rata"
)

type Client interface {
	CreateWorkstation(request WorkstationCreateRequest) error
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
