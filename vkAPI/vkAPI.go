package vkAPI

import (
    "net/url"
    "net/http"
    "bytes"
    "io/ioutil"
    "encoding/json"
    "os"
    "log"
    "strconv"
)

const VKAuthUrl = "https://oauth.vk.com/authorize?"
const VKApiUrl = "https://api.vk.com/method/"
const VKApiVesrsion = "5.67"

type Method struct {
    object string
    action string
}

func accessToken() string {
    return os.Getenv("VKUSERTOKEN")
}

type HttpResponse struct {
    response *http.Response
    err      error
}

func (m Method) EncodeUrlParams(apiUrl string, params *url.Values) (string, error) {
    u, err := url.Parse(apiUrl)
    if err != nil {
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

func GetWallPosts(owner_id int, offset int, count int, filter string, extendedFields string) (WallItems, error) {
    params := url.Values{}
    params.Add("access_token", accessToken())
    params.Add("owner_id", strconv.Itoa(-owner_id))
    params.Add("offset", strconv.Itoa(offset))
    params.Add("count", strconv.Itoa(count))
    params.Add("v", "5.67")
    params.Add("filter", filter)
    if extendedFields != "" {
        params.Add("extended", "1")
        params.Add("extended_fields", extendedFields)
    } else {
        params.Add("extended", "0")
    }
    httpResponse := <-Request(Method{"wall", ".get"}, &params, true)
    body, err := ioutil.ReadAll(httpResponse.response.Body)
    if err != nil {
        return WallItems{}, err
    }
    var wallResponse WallResponse
    err = json.Unmarshal(body, &wallResponse)
    return wallResponse.Response.Items, err
}


func Request(method Method, params *url.Values, isAccessTokenRequired bool) chan *HttpResponse {
    reqUrl, _ := method.EncodeUrlParams(VKApiUrl, params)
    params.Add("v", VKApiVesrsion)
    if isAccessTokenRequired {
        params.Add("access_token", accessToken())
    }
    response := make(chan *HttpResponse, 1)
    go func(u string, ch chan *HttpResponse) {
        log.Println(u)
        r, err := http.Get(u)
        response <- &HttpResponse{r, err}
    }(reqUrl, response)
    return response
}


