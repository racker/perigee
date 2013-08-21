// vim: ts=8 sw=8 noet ai

package perigee

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// The UnexpectedResponseCodeError structure represents a mismatch in understanding between server and client in terms of response codes.
// Most often, this is due to an actual error condition (e.g., getting a 404 for a resource when you expect a 200).
// However, it needn't always be the case (e.g., getting a 204 (No Content) response back when a 200 is expected).
type UnexpectedResponseCodeError struct {
	Expected []int
	Actual   int
}

func (err *UnexpectedResponseCodeError) Error() string {
	return fmt.Sprintf("Expected HTTP response code %d; got %d instead", err.Expected, err.Actual)
}

// request is the procedure that does the ditch-work of making the request, marshaling parameters, and unmarshaling results.
func request(method string, url string, opts Options) (*Response, error) {
	var body io.Reader
	var response Response

	acceptableResponseCodes := opts.OkCodes
	if len(acceptableResponseCodes) == 0 {
		acceptableResponseCodes = []int{200}
	}

	client := opts.CustomClient
	if client == nil {
		client = new(http.Client)
	}

	body = nil
	if opts.ReqBody != nil {
		bodyText, err := json.Marshal(opts.ReqBody)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(string(bodyText))
		if opts.DumpReqJson {
			log.Printf("Making request:\n%#v\n", string(bodyText))
		}
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	if opts.MoreHeaders != nil {
		for k, v := range opts.MoreHeaders {
			req.Header.Add(k, v)
		}
	}

	httpResponse, err := client.Do(req)
	response.HttpResponse = *httpResponse
	response.StatusCode = httpResponse.StatusCode
	defer httpResponse.Body.Close()

	if err != nil {
		return &response, err
	}
	// This if-statement is legacy code, preserved for backward compatibility.
	if opts.StatusCode != nil {
		*opts.StatusCode = httpResponse.StatusCode
	}
	if not_in(httpResponse.StatusCode, acceptableResponseCodes) {
		return &response, &UnexpectedResponseCodeError{
			Expected: acceptableResponseCodes,
			Actual:   httpResponse.StatusCode,
		}
	}
	if opts.Results != nil {
		jsonResult, err := ioutil.ReadAll(httpResponse.Body)
		response.JsonResult = jsonResult
		if err != nil {
			return &response, err
		}

		err = json.Unmarshal(jsonResult, opts.Results)
		// This if-statement is legacy code, preserved for backward compatibility.
		if opts.ResponseJson != nil {
			*opts.ResponseJson = jsonResult
		}
	}
	return &response, err
}

// not_in returns false if, and only if, the provided needle is _not_
// in the given set of integers.
func not_in(needle int, haystack []int) bool {
	for _, straw := range haystack {
		if needle == straw {
			return false
		}
	}
	return true
}

// Post makes a POST request against a server using the provided HTTP client.
// The url must be a fully-formed URL string.
func Post(url string, opts Options) error {
	_, err := request("POST", url, opts)
	return err
}

// Get makes a GET request against a server using the provided HTTP client.
// The url must be a fully-formed URL string.
func Get(url string, opts Options) error {
	_, err := request("GET", url, opts)
	return err
}

// Delete makes a DELETE request against a server using the provided HTTP client.
// The url must be a fully-formed URL string.
func Delete(url string, opts Options) error {
	_, err := request("DELETE", url, opts)
	return err
}

// Put makes a PUT request against a server using the provided HTTP client.
// The url must be a fully-formed URL string.
func Put(url string, opts Options) error {
	_, err := request("PUT", url, opts)
	return err
}

// Options describes a set of optional parameters to the various request calls.
//
// The custom client can be used for a variety of purposes beyond selecting encrypted versus unencrypted channels.
// Transports can be defined to provide augmented logging, header manipulation, et. al.
//
// If the ReqBody field is provided, it will be embedded as a JSON object.
// Otherwise, provide nil.
//
// If JSON output is to be expected from the response,
// provide either a pointer to the container structure in Results,
// or a pointer to a nil-initialized pointer variable.
// The latter method will cause the unmarshaller to allocate the container type for you.
// If no response is expected, provide a nil Results value.
//
// The MoreHeaders map, if non-nil or empty, provides a set of headers to add to those
// already present in the request.  At present, only Accepted and Content-Type are set
// by default.
//
// OkCodes provides a set of acceptable, positive responses.
//
// If provided, StatusCode specifies a pointer to an integer, which will receive the
// returned HTTP status code, successful or not.  DEPRECATED; use the Response.StatusCode field instead for new software.
//
// ResponseJson, if specified, provides a means for returning the raw JSON.  This is
// most useful for diagnostics.  DEPRECATED; use the Response.JsonResult field instead for new software.
//
// DumpReqJson, if set to true, will cause the request to appear to stdout for debugging purposes.
// This attribute may be removed at any time in the future; DO NOT use this attribute in production software.
type Options struct {
	CustomClient *http.Client
	ReqBody      interface{}
	Results      interface{}
	MoreHeaders  map[string]string
	OkCodes      []int
	StatusCode   *int `DEPRECATED`
	DumpReqJson  bool `UNSUPPORTED`
	ResponseJson *[]byte `DEPRECATED`
}

// Response contains return values from the various request calls.
//
// HttpResponse will return the http response from the request call.
// Note: HttpResponse.Body is always closed and will not be available from this return value.
//
// StatusCode specifies the returned HTTP status code, successful or not.
//
// If Results is specified in the Options:
// - JsonResult will contain the raw return from the request call
//   This is most useful for diagnostics.
// - Result will contain the unmarshalled json either in the Result passed in
//   or the unmarshaller will allocate the container type for you.

type Response struct {
  HttpResponse http.Response
  JsonResult   []byte
  Results      interface{}
  StatusCode   int
}