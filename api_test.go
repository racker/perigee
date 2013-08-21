package perigee

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormal(t *testing.T) {
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("testing"))
		})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	response, err := request("GET", ts.URL, Options{})
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("response code %d is not 200", response.StatusCode)
	}
}

func TestOKCodes(t *testing.T) {
	expectCode := 201
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(expectCode)
			w.Write([]byte("testing"))
		})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	options := Options{
		OkCodes: []int{expectCode},
	}
	results, err := request("GET", ts.URL, options)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if results.StatusCode != expectCode {
		t.Fatalf("response code %d is not %d", results.StatusCode, expectCode)
	}
}

func TestLocation(t *testing.T) {
	newLocation := "http://www.example.com"
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Location", newLocation)
			w.Write([]byte("testing"))
		})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	response, err := request("GET", ts.URL, Options{})
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	location, err := response.HttpResponse.Location()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if location.String() != newLocation {
		t.Fatalf("location returned \"%s\" is not \"%s\"", location.String(), newLocation)
	}
}

func TestHeaders(t *testing.T) {
	newLocation := "http://www.example.com"
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Location", newLocation)
			w.Write([]byte("testing"))
		})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	response, err := request("GET", ts.URL, Options{})
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	location := response.HttpResponse.Header.Get("Location")
	if location == "" {
		t.Fatalf("Location should not empty")
	}

	if location != newLocation {
		t.Fatalf("location returned \"%s\" is not \"%s\"", location, newLocation)
	}
}

func TestJson(t *testing.T) {
	newLocation := "http://www.example.com"
	jsonBytes := []byte(`{"foo": {"bar": "baz"}}`)
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Location", newLocation)
			w.Write(jsonBytes)
		})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	type Data struct {
		Foo struct {
			Bar string `json:"bar"`
		} `json:"foo"`
	}
	var data Data

	response, err := request("GET", ts.URL, Options{Results: &data})
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if bytes.Compare(jsonBytes, response.JsonResult) != 0 {
		t.Fatalf("json returned \"%s\" is not \"%s\"", response.JsonResult, jsonBytes)
	}

	if data.Foo.Bar != "baz" {
		t.Fatalf("Results returned %v", data)
	}
}
