// vim: ts=8 sw=8 noet ai

package perigee


import (
	"fmt"
	"net/http"
	"io"
	"io/ioutil"
	"encoding/json"
	"strings"
)


// The UnexpectedResponseCodeError structure represents a mismatch in understanding between server and client in terms of response codes.
// Most often, this is due to an actual error condition (e.g., getting a 404 for a resource when you expect a 200).
// However, it needn't always be the case (e.g., getting a 204 (No Content) response back when a 200 is expected).
type UnexpectedResponseCodeError struct {
	Expected, Actual int
}

func (err *UnexpectedResponseCodeError) Error() string {
	return fmt.Sprintf("Expected HTTP response code %d; got %d instead", err.Expected, err.Actual)
}


// request is the procedure that does the ditch-work of making the request, marshaling parameters, and unmarshaling results.
func request(method string, url string, opts Options) error {
	var body io.Reader

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
	if response.StatusCode != 200 {
		return &UnexpectedResponseCodeError{
			Expected: 200,
			Actual: response.StatusCode,
		}
	}
	defer response.Body.Close()

	if opts.Results != nil {
		jsonResult, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(jsonResult, opts.Results)
	}
	return err
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
type Options struct {
	CustomClient *http.Client
	ReqBody interface{}
	Results interface{}
	MoreHeaders map[string]string
}

