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

func NewBotApi(token string) *BotApi {
	return &BotApi{Token: token}
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

func (api *BotApi) EncodeDownloadUrl(filePath string) string {
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

	response, err := api.SendGetToApi("getFile", &params)
	defer response.Body.Close()
	if err != nil {
		//TODO: Parse error
		return File{}, err
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		//TODO: determine what kind of error it could be and handle it
		return File{}, err
	}

	var result GetFileResponse
	err = json.Unmarshal(body, &result)

	if err != nil {
		return File{}, err
	}
	return result.File, nil
}

func (api *BotApi) SetWebhook(url string, certPath string, maxConn int, allowedUpdates string) error {
	reqBody, writer := api.createWebhookRequestBody(url, maxConn, allowedUpdates)
	defer writer.Close()
	if certPath != "" {
		err := api.loadLocalCertificate(certPath, reqBody, writer)
		if err != nil {
			return err
		}
	}

	contentType := writer.FormDataContentType()
	resp, err := api.SendPostToApi("setWebhook", contentType, reqBody)
	defer resp.Body.Close()
	return err
}

func (api *BotApi) createWebhookRequestBody(url string, maxConn int, allowedUpdates string) (*bytes.Buffer, *multipart.Writer) {
	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	bodyWriter.WriteField("allowed_updates", allowedUpdates)
	bodyWriter.WriteField("max_connections", strconv.Itoa(maxConn))
	bodyWriter.WriteField("url", url)

	return bodyBuffer, bodyWriter
}

func (api *BotApi) loadLocalCertificate(certPath string, buff *bytes.Buffer, wr *multipart.Writer) error {
	fileWriter, err := wr.CreateFormFile("certificate", certPath)
	if err != nil {
		return err
	}

	file, err := os.Open(certPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(fileWriter, file)
	return err
}

func (api *BotApi) SendMessage(chatId uint64, text string, withoutPreview bool) (*http.Response, error) {
	message := &SendMessageRequest{
		Text:                  text,
		ChatId:                chatId,
		DisableWebPagePreview: withoutPreview,
	}
	var buff bytes.Buffer
	mjson, err := json.Marshal(&message)
	if err != nil {
		return nil, err
	}
	buff.Write(mjson)
	r, err := api.SendPostToApi("sendMessage", "application/json", &buff)
	return r, err
}

func (api *BotApi) buildApiUrl(method string) *bytes.Buffer {
	var buff bytes.Buffer
	buff.WriteString(TGApiUrl)
	buff.WriteString(api.Token)
	buff.WriteString("/")
	buff.WriteString(method)
	return &buff
}
