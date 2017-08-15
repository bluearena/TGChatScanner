package vkAPI

import (
    "net/url"
    "log"
    "runtime"
    "os/exec"
    "fmt"
    "strconv"
    "os"
)
func redirectUri() string{
    return os.Getenv("REDIRECTURI")
}

func OAuthAuthorize(redirectUri string, scope ...int, ) error {
    params := url.Values{}
    params.Add("client_id", "6148845")
    params.Add("redirect_uri", redirectUri)
    encodedMask := encodeScope(scope...)
    log.Println(encodedMask)
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

func ImplicitFlow(scope ...int) error {
    return OAuthAuthorize("https://oauth.vk.com/blank.html", scope...)
}

func AuthorizationCodeFlow(scope ...int) error {
    return OAuthAuthorize( redirectUri(), scope...)
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
    for _, val := range scope {
        total += val
    }

    return strconv.Itoa(total)
}
