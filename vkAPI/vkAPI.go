package vkAPI

import (
    "net/url"
    "os/exec"
    "runtime"
    "fmt"
    "net/http"
    "bytes"
    "io/ioutil"
    "encoding/json"
    "os"
    "log"
    "strconv"
)

const VKAuthUrl= "https://oauth.vk.com/authorize?"
const VKApiUrl = "https://api.vk.com/method/"


type Method struct {
    object string
    action string
}

func accessToken() string{
    return os.Getenv("VKUSERTOKEN")
}

func (m Method) Encode(apiUrl string, params *url.Values) (string, error){
    u, err := url.Parse(apiUrl)
    if err != nil{
        return "", err
    }
    var buffer bytes.Buffer
    buffer.WriteString(u.String())
    buffer.WriteString(m.object)
    buffer.WriteString(m.action)
    buffer.WriteString("?")
    buffer.WriteString(params.Encode())
    return buffer.String(), nil
}

func GetWallPosts(owner_id int, offset int, count int, filter string, extendedFields string) (WallItems, error){
    params := url.Values{}
    params.Add("access_token", accessToken())
    params.Add("owner_id", strconv.Itoa(-owner_id))
    params.Add("offset", strconv.Itoa(offset))
    params.Add("count",strconv.Itoa(count))
    params.Add("v","5.67")
    params.Add("filter",filter)
    if extendedFields != ""{
        params.Add("extended", "1")
        params.Add("extended_fields",extendedFields)
    }else{
        params.Add("extended", "0")
    }
    response := <-Request(Method{"wall",".get"}, &params)
    body, err := ioutil.ReadAll(response.Body)
    if err != nil{
        return WallItems{}, err
    }
    var wallResponse WallResponse
    err = json.Unmarshal(body,&wallResponse)
    return wallResponse.Response.Items, err
}

func Request(method Method, params *url.Values) chan *http.Response {
    reqUrl, _ := method.Encode(VKApiUrl, params)
    response := make(chan *http.Response, 1)
    go func (u string, ch chan *http.Response){
        log.Println(u)
        r, err := http.Get(u)
        if err != nil {
            log.Println(err)
        }
        response <- r
    }(reqUrl, response)
    return response
}

func UserImplicitFlow(scope ...int) error {
    params := url.Values{}
    params.Add("client_id", "6148845")
    params.Add("redirect_uri", "https://oauth.vk.com/blank.html")
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
    for _, val := range scope {
        total += val
    }

    return strconv.Itoa(total)
}
