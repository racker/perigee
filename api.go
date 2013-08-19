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
func request(method string, url string, opts Options) error {
	var body io.Reader

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
			return err
		}
		body = strings.NewReader(string(bodyText))
		if opts.DumpReqJson {
			log.Printf("Making request:\n%#v\n", string(bodyText))
		}
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	if opts.MoreHeaders != nil {
		for k, v := range opts.MoreHeaders {
			req.Header.Add(k, v)
		}
	}

	response, err := client.Do(req)
	if err != nil {
		return err
	}
	if opts.Location != nil {
		location, err := response.Location()
		if err == nil {
			*opts.Location = location.String()
		}
	}
	if opts.StatusCode != nil {
		*opts.StatusCode = response.StatusCode
	}
	if not_in(response.StatusCode, acceptableResponseCodes) {
		return &UnexpectedResponseCodeError{
			Expected: acceptableResponseCodes,
			Actual:   response.StatusCode,
		}
	}
	defer response.Body.Close()
	if opts.Results != nil {
		jsonResult, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(jsonResult, opts.Results)
		if opts.ResponseJson != nil {
			*opts.ResponseJson = jsonResult
		}
	}
	return err
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
	return request("POST", url, opts)
}

// Get makes a GET request against a server using the provided HTTP client.
// The url must be a fully-formed URL string.
func Get(url string, opts Options) error {
	return request("GET", url, opts)
}

// Delete makes a DELETE request against a server using the provided HTTP client.
// The url must be a fully-formed URL string.
func Delete(url string, opts Options) error {
	return request("DELETE", url, opts)
}

// Put makes a PUT request against a server using the provided HTTP client.
// The url must be a fully-formed URL string.
func Put(url string, opts Options) error {
	return request("PUT", url, opts)
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
// returned HTTP status code, successful or not.
//
// ResponseJson, if specified, provides a means for returning the raw JSON.  This is
// most useful for diagnostics.
type Options struct {
	CustomClient *http.Client
	ReqBody      interface{}
	Results      interface{}
	MoreHeaders  map[string]string
	OkCodes      []int
	StatusCode   *int
	Location     *string
	DumpReqJson  bool
	ResponseJson *[]byte
}
