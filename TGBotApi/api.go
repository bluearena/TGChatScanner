package TGBotApi

import (
    "net/url"
    "os"
    "net/http"
    "bytes"

    "strings"
    "io/ioutil"
    "encoding/json"
)

const (
    TGApiUrl = "https://api.telegram.org/bot"
)


func encodeApiUrl(method string, params *url.Values) string{
   var buff bytes.Buffer
    buff.WriteString(TGApiUrl)
    buff.WriteString(botToken())
    buff.WriteString("/")
    buff.WriteString(method)
    buff.WriteString("?")
    buff.WriteString(params.Encode())
    return buff.String()
}

func sendRequestToApi(method string, params *url.Values) (*http.Response, error){
    reqUrl := encodeApiUrl(method,params)
    response, err := http.Get(reqUrl)
    return response,err
}

func botToken() string{
    return os.Getenv("BOTACCESS")
}

func PrepareFile(fileId string) (File, error){
    params:= url.Values{}
    params.Add("file_id", fileId)
    response, err := sendRequestToApi("upload.GetFile", &params)
    if err != nil{
        //TODO: Parse error
        return File{}, err
    }
    body, err := ioutil.ReadAll(response.Body)
    if err != nil{
        //TODO: determine what kind of error it could be and handle it
    }
    var result File
    err = json.Unmarshal(body,&result)
    if err != nil{
        return File{}, err
    }
    return result, nil
}







