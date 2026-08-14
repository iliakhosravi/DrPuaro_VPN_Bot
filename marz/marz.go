package marz

import (
	"fmt"
	"sync"

	"github.com/go-resty/resty/v2"
	"techybat.org/go-vpn/vars"
)

const (
	LOGIN_PATH = "/api/admins/token"
)

type Marz struct {
	client *resty.Client
}

var (
	marz     *Marz
	marzOnce sync.Once
)

func GetMarz() *Marz {
	marzOnce.Do(func() {
		Setup()
	})

	return marz
}

func Setup() {
	url := vars.Get("MARZ_API_URL")
	marz = &Marz{
		client: resty.New().SetBaseURL(url),
	}

	if err := marz.Login(vars.Get("MARZ_API_USERNAME"), vars.Get("MARZ_API_PASSWORD")); err != nil {
		panic(err)
	}
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

func (mz *Marz) Login(username, password string) error {
	var res tokenResponse
	resp, err := mz.client.R().
		SetFormData(map[string]string{
			"username":   username,
			"password":   password,
			"grant_type": "password",
		}).
		SetResult(&res).
		Post(LOGIN_PATH)

	if err != nil {
		return fmt.Errorf("unable to login to marzneshin panel. Err: %s", err)
	}

	if resp.IsError() {
		return fmt.Errorf("marzneshin panel login failed. Response from server: %s", resp.String())
	}

	mz.client.SetAuthToken(res.AccessToken)
	return nil
}
