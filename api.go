// vim: ts=8 sw=8 noet ai

package perigee


import (
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

func (err *UnexpectedResponseError) Error() string {
	return fmt.Sprintf("Expected HTTP response code %d; got %d instead", err.Expected, err.Actual)
}


// Post makes a POST request against a server using the provided HTTP client.
// The url must be a fully-formed URL string.
// If input is to be embedded in a POST request body, it will be encoded in JSON format.
// Otherwise, provide nil for input.
// If output is to be expected from the response,
// provide either a pointer to the container,
// or a pointer to a nil-initialized pointer variable.
// The latter method will cause Post to allocate the container type for you.
// If no response is expected, provide a nil output reference.
func Post(client *http.Client, url string, input interface{}, output interface{}) error {
	var body io.Reader

	body = nil
	if input != nil {
		bodyText, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = strings.NewReader(string(bodyText))
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return &UnexpectedResponseCodeError{
			Expected: 200,
			Actual: response.StatusCode
		}
	}
	defer response.Body.Close()

	if output != nil {
		jsonResult, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(jsonResult, output)
	}
	return err
}

