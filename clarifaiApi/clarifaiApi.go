package clarifaiApi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const apiUrl = "https://api.clarifai.com/v2/models/aaa03c23b3724a16a56b629203edc62c/outputs"
const ApiKey = "aaa03c23b3724a16a56b629203edc62c"

type ClarifaiApi struct {
	httpClient *http.Client
	apiKey     string
}

var api *ClarifaiApi = nil

func GetApiInstance() *ClarifaiApi{
	if api != nil{
		return api
	}
	return NewClarifaiApi(apiUrl)
}

func NewClarifaiApi(apiKey string) *ClarifaiApi {
	return &ClarifaiApi{
		httpClient: &http.Client{},
		apiKey:     apiKey,
	}
}

func (ca *ClarifaiApi) RecognizeImage(url string, minProbability float64) ([]string, error) {
	result, err := ca.RecognizeImages([]string{url}, minProbability)

	if err != nil {
		return nil, err
	}

	return result[0], nil
}

// Recognize images and return tags
func (ca *ClarifaiApi) RecognizeImages(urls []string, minProbability float64) ([][]string, error) {
	requestBody := Request{make([]Input, len(urls))}

	for i, url := range urls {
		requestBody.Inputs[i] = Input{Data: Data{Image{url}}}
	}

	b, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Key "+ca.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := ca.httpClient.Do(req)
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	responseBody := new(Response)

	err = json.Unmarshal(body, responseBody)
	if err != nil {
		fmt.Println("error:", err)
	}

	tags := make([][]string, len(responseBody.Outputs))
	for i, output := range responseBody.Outputs {
		if output.Status.Code != 10000 {
			err := errors.New(output.Input.Data.Image.Url + ": " + output.Status.Description)
			return nil, err
		}

		for _, concept := range output.ConceptData.Concepts {
			if concept.Value >= minProbability {
				tags[i] = append(tags[i], concept.Name)
			}
		}
	}

	return tags, nil
}
