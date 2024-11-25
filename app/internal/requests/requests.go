package requests

import (
	"fmt"
	"net/http"
)

func Get(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d != 200", resp.StatusCode)
	}
	return resp, nil
}
