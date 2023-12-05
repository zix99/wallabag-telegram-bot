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

func (s *Wallabag) AddURL(bagUrl, tags string) (int, error) {
	if s.token == "" {
		token, err := s.getNewToken()
		if err != nil {
			return -1, err
		}
		s.token = token
	}

	data := map[string]any{
		"url":  bagUrl,
		"tags": tags,
	}
	dataBytes, _ := json.Marshal(data)

	req, err := http.NewRequest(http.MethodPost, s.Url+"/api/entries.json", bytes.NewReader(dataBytes))
	if err != nil {
		return -1, err
	}

	req.Header.Add("Authorization", "Bearer "+s.token)
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	if res.StatusCode/100 != 2 {
		logrus.Warn(string(body))
		return -1, fmt.Errorf("Unexpected status: %d", res.StatusCode)
	}

	respObj := struct {
		Id int
	}{}
	json.Unmarshal(body, &respObj)

	return respObj.Id, nil
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
