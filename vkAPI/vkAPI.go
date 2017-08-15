package vkAPI

import (
	"net/url"
	"os/exec"
	"runtime"
	"fmt"
)

const VKAuthUrl = "https://oauth.vk.com/authorize?"

func UserImplicitFlow(scope ...int) error {
	params := url.Values{}
	params.Add("client_id", "6148845")
	params.Add("redirect_uri", "https://oauth.vk.com/blank.html")
	encodedMask := encodeScope(scope...)
	params.Add("scope", encodedMask)
	params.Add("response_type", "token")
	params.Add("display", "page")
	authUrl := VKAuthUrl + params.Encode()
	err := browserOpen(authUrl)
	if err != nil {
		return err
	}
	return nil
}

func GroupImplicitFlow(userID int, scope ...int) error {
	params := url.Values{}
	params.Add("client_id", "6148845")
	params.Add("redirect_uri", "https://oauth.vk.com/blank.html")
	params.Add("response_type", "token")
	params.Add("display", "page")
	encodedMask := encodeScope(scope...)
	params.Add("scope", encodedMask)
	err := browserOpen(VKAuthUrl + params.Encode())
	return err
}

func browserOpen(URL string) (err error) {
	err = nil
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", URL).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", URL).Start()
	case "darwin":
		err = exec.Command("open", URL).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

func encodeScope(scope ...int) string {
	total := 0
	for i := range scope {
		total += i
	}
	return string(total)
}
