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

func BuildGetUrl(method string, params *url.Values) string {
	buff := buildApiUrl(method)
	buff.WriteString("?")
	buff.WriteString(params.Encode())
	return buff.String()
}

func EncodrApiUrl(method string) string {
	buff := buildApiUrl(method)
	return buff.String()
}

func EncodeDownloadUrl(filePath string) string {
	var buff bytes.Buffer
	buff.WriteString(TGDownloadUrl)
	buff.WriteString(botToken())
	buff.WriteString("/")
	buff.WriteString(filePath)
	return buff.String()
}

func SendGetToApi(method string, params *url.Values) (*http.Response, error) {
	reqUrl := BuildGetUrl(method, params)
	response, err := http.Get(reqUrl)
	return response, err
}

func SendPostToApi(method string, contentType string, buffer *bytes.Buffer) (*http.Response, error) {
	reqUrl := EncodrApiUrl(method)
	response, err := http.Post(reqUrl, contentType, buffer)
	return response, err
}

func PrepareFile(fileId string) (File, error) {
	params := url.Values{}
	params.Add("file_id", fileId)

	response, err := SendGetToApi("upload.GetFile", &params)
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

func SetWebhook(url string, certPath string, maxConn int, allowedUpdates string) error {

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

	resp, err := SendPostToApi("setWebhook", contentType, bodyBuffer)
	defer resp.Body.Close()
	return err
}

func GetWebhookUrl() string {
	return "/" + botToken()
}

func botToken() string {
	return os.Getenv("BOTACCESS")
}

func buildApiUrl(method string) *bytes.Buffer {
	var buff bytes.Buffer
	buff.WriteString(TGApiUrl)
	buff.WriteString(botToken())
	buff.WriteString("/")
	buff.WriteString(method)
	return &buff
}
