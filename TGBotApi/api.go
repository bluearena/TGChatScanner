package TGBotApi

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const (
	TGApiUrl      = "https://api.telegram.org/bot"
	TGDownloadUrl = "https://api.telegram.org/file/bot"
)

type BotApi struct {
	Token string
}


func NewBotApi(token string) *BotApi{
	return &BotApi{Token : token}
}


func (api *BotApi) BuildGetUrl(method string, params *url.Values) string {
	buff := api.buildApiUrl(method)
	buff.WriteString("?")
	buff.WriteString(params.Encode())
	return buff.String()
}

func (api *BotApi) EncodrApiUrl(method string) string {
	buff := api.buildApiUrl(method)
	return buff.String()
}

func(api *BotApi) EncodeDownloadUrl(filePath string) string {
	var buff bytes.Buffer
	buff.WriteString(TGDownloadUrl)
	buff.WriteString(api.Token)
	buff.WriteString("/")
	buff.WriteString(filePath)
	return buff.String()
}

func (api *BotApi) SendGetToApi(method string, params *url.Values) (*http.Response, error) {
	reqUrl := api.BuildGetUrl(method, params)
	response, err := http.Get(reqUrl)
	return response, err
}

func (api *BotApi) SendPostToApi(method string, contentType string, buffer *bytes.Buffer) (*http.Response, error) {
	reqUrl := api.EncodrApiUrl(method)
	response, err := http.Post(reqUrl, contentType, buffer)
	return response, err
}

func (api *BotApi) PrepareFile(fileId string) (File, error) {
	params := url.Values{}
	params.Add("file_id", fileId)

	response, err := api.SendGetToApi("upload.GetFile", &params)
	if err != nil {
		//TODO: Parse error
		return File{}, err
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		//TODO: determine what kind of error it could be and handle it
	}

	var result File
	err = json.Unmarshal(body, &result)
	if err != nil {
		return File{}, err
	}
	return result, nil
}

func (api *BotApi) SetWebhook(url string, certPath string, maxConn int, allowedUpdates string) error {

	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	fileWriter, err := bodyWriter.CreateFormFile("certificate", certPath)

	if err != nil {
		bodyWriter.Close()
		return err
	}

	file, err := os.Open(certPath)
	defer file.Close()

	if err != nil {
		bodyWriter.Close()
		return err
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		bodyWriter.Close()
		return err
	}

	bodyWriter.WriteField("allowed_updates", allowedUpdates)
	bodyWriter.WriteField("max_connections", strconv.Itoa(maxConn))
	bodyWriter.WriteField("url", url)

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := api.SendPostToApi("setWebhook", contentType, bodyBuffer)
	defer resp.Body.Close()
	return err
}

func (api *BotApi) buildApiUrl(method string) *bytes.Buffer {
	var buff bytes.Buffer
	buff.WriteString(TGApiUrl)
	buff.WriteString(api.Token)
	buff.WriteString("/")
	buff.WriteString(method)
	return &buff
}
