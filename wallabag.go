package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

type Wallabag struct {
	Url                    string
	ClientID, ClientSecret string
	Username, Password     string

	token string
}

func (s *Wallabag) AddURL(bagUrl string) error {
	if s.token == "" {
		token, err := s.getNewToken()
		if err != nil {
			return err
		}
		s.token = token
	}

	// form := url.Values{}
	// form.Add("url", bagUrl)
	//form.Add("tags", "telegram")
	data := map[string]any{
		"url":  bagUrl,
		"tags": "telegram",
	}
	dataBytes, _ := json.Marshal(data)

	req, err := http.NewRequest(http.MethodPost, s.Url+"/api/entries.json", bytes.NewReader(dataBytes))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+s.token)
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	if res.StatusCode/100 != 2 {
		logrus.Warn(string(body))
		return fmt.Errorf("Unexpected status: %d", res.StatusCode)
	}

	return nil
}

func (s *Wallabag) Test() error {
	tok, err := s.getNewToken()
	logrus.Infof("Got token: %s", tok)
	return err
}

func (s *Wallabag) getNewToken() (string, error) {
	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("client_id", s.ClientID)
	form.Add("client_secret", s.ClientSecret)
	form.Add("username", s.Username)
	form.Add("password", s.Password)

	res, err := http.PostForm(s.Url+"/oauth/v2/token", form)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	retMap := make(map[string]any)
	if err := json.Unmarshal(body, &retMap); err != nil {
		return "", err
	}

	return retMap["access_token"].(string), nil
}
