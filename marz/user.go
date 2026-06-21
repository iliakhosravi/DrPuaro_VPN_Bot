package marz

import "fmt"

const (
	ONE_GB = 1073741824
	ONE_MB = 1048576

	ResetStrategyNoReset = "no_reset"

	ExpireStrategyNever     = "never"
	ExpireStrategyFixedDate = "fixed_date"

	USERS_PATH    = "/api/users"
	SERVICES_PATH = "/api/services"
)

type User struct {
	ID                     int     `json:"id,omitempty"`
	Username               string  `json:"username"`
	Note                   string  `json:"note,omitempty"`
	DataLimit              int64   `json:"data_limit"`
	DataLimitResetStrategy string  `json:"data_limit_reset_strategy"`
	ExpireStrategy         string  `json:"expire_strategy"`
	ExpireDate             *string `json:"expire_date,omitempty"`
	ServiceIDs             []int   `json:"service_ids"`
	Enabled                bool    `json:"enabled,omitempty"`
	Expired                bool    `json:"expired,omitempty"`
	DataLimitReached       bool    `json:"data_limit_reached,omitempty"`
	UsedTraffic            int64   `json:"used_traffic,omitempty"`
	SubscriptionURL        string  `json:"subscription_url,omitempty"`
}

type Service struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	InboundIDs []int  `json:"inbound_ids"`
}

type servicesResponse struct {
	Items []Service `json:"items"`
}

func (mz *Marz) GetUser(username string) (*User, error) {
	var user User
	resp, err := mz.client.R().
		SetResult(&user).
		Get(USERS_PATH + "/" + username)

	if err != nil {
		return nil, fmt.Errorf("error: getting marzneshin user failed.\nerr:%s", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("error: marzneshin panel returned no user. response: %s", resp.String())
	}

	return &user, nil
}

func (mz *Marz) AddUser(user User) (*User, error) {
	var res User
	resp, err := mz.client.R().
		SetBody(user).
		SetResult(&res).
		Post(USERS_PATH)

	if err != nil {
		return nil, fmt.Errorf("error: adding marzneshin user failed.\nerr:%s", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("error: marzneshin panel couldn't add user. response: %s", resp.String())
	}

	return &res, nil
}

func (mz *Marz) UpdateUser(username string, user User) (*User, error) {
	var res User
	resp, err := mz.client.R().
		SetBody(user).
		SetResult(&res).
		Put(USERS_PATH + "/" + username)

	if err != nil {
		return nil, fmt.Errorf("error: updating marzneshin user failed.\nerr:%s", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("error: marzneshin panel couldn't update user. response: %s", resp.String())
	}

	return &res, nil
}

func (mz *Marz) StoreUser(user User) (*User, error) {
	existing, err := mz.GetUser(user.Username)
	if err != nil || existing == nil {
		return mz.AddUser(user)
	}
	return mz.UpdateUser(user.Username, user)
}

func (mz *Marz) ResetUser(username string) (*User, error) {
	var res User
	resp, err := mz.client.R().
		SetResult(&res).
		Post(USERS_PATH + "/" + username + "/reset")

	if err != nil {
		return nil, fmt.Errorf("error: reset marzneshin user failed.\nerr:%s", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("error: marzneshin panel couldn't reset user. response: %s", resp.String())
	}

	return &res, nil
}

func (mz *Marz) GetServices(page, size int) ([]Service, error) {
	var res servicesResponse
	resp, err := mz.client.R().
		SetQueryParams(map[string]string{
			"page": fmt.Sprint(page),
			"size": fmt.Sprint(size),
		}).
		SetResult(&res).
		Get(SERVICES_PATH)

	if err != nil {
		return nil, fmt.Errorf("error: getting marzneshin services failed.\nerr:%s", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("error: marzneshin panel returned no services. response: %s", resp.String())
	}

	return res.Items, nil
}

func (u User) RemainedTraffic() int {
	if u.DataLimit == 0 {
		return 0
	}
	return int(u.DataLimit-u.UsedTraffic) / ONE_MB
}
