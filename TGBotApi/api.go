package TGBotAPI

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
	TGAPIURL      = "https://API.telegram.org/bot"
	TGDownloadURL = "https://api.telegram.org/file/bot"
)

type BotAPI struct {
	Token string
}

func NewBotAPI(token string) *BotAPI {
	return &BotAPI{Token: token}
}

func (API *BotAPI) BuildGetURL(method string, params *url.Values) string {
	buff := API.buildAPIURL(method)
	buff.WriteString("?")
	buff.WriteString(params.Encode())
	return buff.String()
}

func (API *BotAPI) EncodrAPIURL(method string) string {
	buff := API.buildAPIURL(method)
	return buff.String()
}

func (API *BotAPI) EncodeDownloadURL(filePath string) (string, error) {
	var buff bytes.Buffer
	buff.WriteString(TGDownloadURL)
	buff.WriteString(API.Token)
	buff.WriteString("/")
	buff.WriteString(filePath)
	u, err := url.Parse(buff.String())
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (API *BotAPI) SendGetToAPI(method string, params *url.Values) (*http.Response, error) {
	reqURL := API.BuildGetURL(method, params)
	response, err := http.Get(reqURL)
	return response, err
}

func (API *BotAPI) SendPostToAPI(method string, contentType string, buffer *bytes.Buffer) (*http.Response, error) {
	reqURL := API.EncodrAPIURL(method)
	response, err := http.Post(reqURL, contentType, buffer)
	return response, err
}

func (API *BotAPI) PrepareFile(fileId string) (File, error) {
	params := url.Values{}
	params.Add("file_id", fileId)

	response, err := API.SendGetToAPI("getFile", &params)
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

func (API *BotAPI) SetWebhook(URL string, certPath string, maxConn int, allowedUpdates string) error {
	reqBody, writer := API.createWebhookRequestBody(URL, maxConn, allowedUpdates)
	defer writer.Close()
	if certPath != "" {
		err := API.loadLocalCertificate(certPath, reqBody, writer)
		if err != nil {
			return err
		}
	}

	contentType := writer.FormDataContentType()
	resp, err := API.SendPostToAPI("setWebhook", contentType, reqBody)
	defer resp.Body.Close()
	return err
}

func (API *BotAPI) createWebhookRequestBody(URL string, maxConn int, allowedUpdates string) (*bytes.Buffer, *multipart.Writer) {
	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	bodyWriter.WriteField("allowed_updates", allowedUpdates)
	bodyWriter.WriteField("max_connections", strconv.Itoa(maxConn))
	bodyWriter.WriteField("url", URL)

	return bodyBuffer, bodyWriter
}

func (API *BotAPI) loadLocalCertificate(certPath string, buff *bytes.Buffer, wr *multipart.Writer) error {
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

func (API *BotAPI) SendMessage(chatId int64, text string, withoutPreview bool) (*http.Response, error) {
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
	r, err := API.SendPostToAPI("sendMessage", "application/json", &buff)
	return r, err
}

func (API *BotAPI) buildAPIURL(method string) *bytes.Buffer {
	var buff bytes.Buffer
	buff.WriteString(TGAPIURL)
	buff.WriteString(API.Token)
	buff.WriteString("/")
	buff.WriteString(method)
	return &buff
}
