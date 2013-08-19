package perigee

import (
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

	var code int

	options := Options{
		StatusCode: &code,
	}
	err := request("GET", ts.URL, options)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if code != 200 {
		t.Fatalf("response code %d is not 200", code)
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

	var code int

	options := Options{
		StatusCode: &code,
		OkCodes:    []int{expectCode},
	}
	err := request("GET", ts.URL, options)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if code != expectCode {
		t.Fatalf("response code %d is not %d", code, expectCode)
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

	var code int
	var location string

	options := Options{
		StatusCode: &code,
		Location:   &location,
	}
	err := request("GET", ts.URL, options)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if location != newLocation {
		t.Fatalf("location returned \"%s\" is not \"%s\"", location, newLocation)
	}
}
