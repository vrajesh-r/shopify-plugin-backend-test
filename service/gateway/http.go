package gateway

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func HTTPFormRequest(method, url string, form url.Values, body interface{}) error {
	// create client
	httpC := &http.Client{}

	// instantiate request
	req, err := http.NewRequest(method, url, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// make request
	res, err := httpC.Do(req)
	if err != nil {
		return err
	}

	// evaluate response
	defer res.Body.Close()
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("HTTP Request Failed | Status: %d | Body: %s | Headers: %+v", res.StatusCode, string(contents), res.Header)
	}
	return nil
}

type HttpError struct {
	error
	Code int
}

func NewHttpError(message string, code int) *HttpError {
	return &HttpError{
		error: errors.New(message),
		Code:  code,
	}
}
