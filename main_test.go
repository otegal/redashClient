package main

import (
	"testing"
)

func TestGetConfig(t *testing.T) {
	Conf := getConfig()

	if Conf.APIKey == "" {
		t.Errorf("got: %v\nwant: not empty string", Conf.APIKey)
	}
	if Conf.BaseURL == "" {
		t.Errorf("got: %v\nwant: not empty string", Conf.BaseURL)
	}

	for _, v := range Conf.Query {
		if v.Params == "" {
			t.Errorf("got: %v\nwant: not empty string", v.Params)
		}
		if v.QueryID == "" {
			t.Errorf("got: %v\nwant: not empty string", v.QueryID)
		}
	}
}
