package bapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	APP_KEY    = "geekr"
	APP_SECRET = "geekr-dev"
)

type AccessToken struct {
	Token string `json:"token"`
}

type API struct {
	URL string
}

func NewAPI(base string) *API {
	return &API{URL: base}
}

func (api *API) getAccessToken(ctx context.Context) (string, error) {
	body, err := api.httpPost(
		ctx,
		"auth",
		map[string]string{
			"app_key":    APP_KEY,
			"app_secret": APP_SECRET,
		},
	)
	if err != nil {
		return "", err
	}
	var accessToken AccessToken
	_ = json.Unmarshal(body, &accessToken)
	return accessToken.Token, nil
}

func (api *API) httpGet(ctx context.Context, path string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s", api.URL, path))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body, nil
}

func (api *API) httpPost(ctx context.Context, path string, params map[string]string) ([]byte, error) {
	data := make(url.Values)
	for key, val := range params {
		data.Add(key, val)
	}
	resp, err := http.PostForm(
		fmt.Sprintf("%s/%s", api.URL, path),
		data,
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body, nil
}

func (api *API) GetTagList(ctx context.Context, name string) ([]byte, error) {
	token, err := api.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	body, err := api.httpGet(ctx, fmt.Sprintf("%s?token=%s&name=%s", "api/v1/tags", token, name))
	if err != nil {
		return nil, err
	}
	return body, nil
}
