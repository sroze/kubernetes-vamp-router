package vamprouter

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"reflect"
	"strings"
)

type Client struct {
	// HTTP is the Client's internal http.Client, handling HTTP requests
	HTTP *http.Client

	// The URL of the Vamp Router API
	URL string

	// Debug mode to dump requests
	Debug bool
}

func (c *Client) Get(v interface{}, path string) error {
	return c.APIReq(v, "GET", path, nil)
}

func (c *Client) Patch(v interface{}, path string, body interface{}) error {
	return c.APIReq(v, "PATCH", path, body)
}

func (c *Client) Post(v interface{}, path string, body interface{}) error {
	return c.APIReq(v, "POST", path, body)
}

func (c *Client) Put(v interface{}, path string, body interface{}) error {
	return c.APIReq(v, "PUT", path, body)
}

func (c *Client) Delete(path string) error {
	return c.APIReq(nil, "DELETE", path, nil)
}

// Generates an HTTP request for but does not perform the request.
//
// The type of body determines how to encode the request:
//
//   nil         no body
//   io.Reader   body is sent verbatim
//   else        body is encoded as application/json
func (c *Client) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	var ctype string
	var rbody io.Reader

	switch t := body.(type) {
	case nil:
	case string:
		rbody = bytes.NewBufferString(t)
	case io.Reader:
		rbody = t
	default:
		v := reflect.ValueOf(body)
		if !v.IsValid() {
			break
		}
		if v.Type().Kind() == reflect.Ptr {
			v = reflect.Indirect(v)
			if !v.IsValid() {
				break
			}
		}

		j, err := json.Marshal(body)
		if err != nil {
			log.Fatal(err)
		}
		rbody = bytes.NewReader(j)
		ctype = "application/json"
	}
	apiURL := strings.TrimRight(c.URL, "/")
	req, err := http.NewRequest(method, apiURL+path, rbody)
	if err != nil {
		return nil, err
	}

	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}

	return req, nil
}

// Sends a Heroku API request and decodes the response into v. As
// described in NewRequest(), the type of body determines how to
// encode the request body. As described in DoReq(), the type of
// v determines how to handle the response body.
func (c *Client) APIReq(v interface{}, meth, path string, body interface{}) error {
	req, err := c.NewRequest(meth, path, body)
	if err != nil {
		return err
	}
	return c.DoReq(req, v)
}

// Submits an HTTP request, checks its response, and deserializes
// the response into v. The type of v determines how to handle
// the response body:
//
//   nil        body is discarded
//   io.Writer  body is copied directly into v
//   else       body is decoded into v as json
//
func (c *Client) DoReq(req *http.Request, v interface{}) error {
	if c.Debug {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Println(err)
		} else {
			os.Stderr.Write(dump)
			os.Stderr.Write([]byte{'\n', '\n'})
		}
	}

	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if c.Debug {
		dump, err := httputil.DumpResponse(res, true)
		if err != nil {
			log.Println(err)
		} else {
			os.Stderr.Write(dump)
			os.Stderr.Write([]byte{'\n'})
		}
	}
	if err = checkResp(res); err != nil {
		return err
	}
	switch t := v.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(t, res.Body)
	default:
		err = json.NewDecoder(res.Body).Decode(v)
	}
	return err
}

// An Error represents an API error.
type Error struct {
	error
	Status string
}

type errorResp struct {
	Status string
}

func checkResp(res *http.Response) error {
	if res.StatusCode/100 != 2 { // 200, 201, 202, etc
		var e errorResp
		err := json.NewDecoder(res.Body).Decode(&e)
		if err != nil {
			return errors.New("Unexpected error: " + res.Status)
		}
		return Error{error: errors.New(e.Status), Status: e.Status}
	}

	return nil
}
