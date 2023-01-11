package bapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
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
	// 实现 HTTP 请求追踪
	url := fmt.Sprintf("%s/%s", api.URL, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	span, newCtx := opentracing.StartSpanFromContext(
		ctx, "HTTP GET: "+api.URL,
		opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
	)
	span.SetTag("url", url)
	_ = opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)
	// 基于新的 ctx 发起 HTTP 请求
	req = req.WithContext(newCtx)
	client := http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	/*resp, err := http.Get(fmt.Sprintf("%s/%s", api.URL, path))
	if err != nil {
		return nil, err
	}*/
	defer resp.Body.Close()
	defer span.Finish()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

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
	if name == "灭霸" {
		panic("毁灭吧")
	}
	body, err := api.httpGet(ctx, fmt.Sprintf("%s?token=%s&name=%s", "api/v1/tags", token, name))
	if err != nil {
		return nil, err
	}
	return body, nil
}
