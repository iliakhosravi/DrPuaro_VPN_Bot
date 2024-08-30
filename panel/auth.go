package panel

import "fmt"

const (
	LOGIN_PATH = "/login"
)

func (panel *Panel) Login(username, password string) error {
	var br BasicResponse
	_, err := panel.client.R().
		SetFormData(
			map[string]string{
				"username": username,
				"password": password,
			},
		).
		SetResult(&br).
		Post(LOGIN_PATH)

	if err != nil {
		return fmt.Errorf("unable to login to panel. Err: %s", err)
	}

	if !br.Success {
		return fmt.Errorf("panel login failed. Response from server: %s", br.Message)
	}
	return nil
}
