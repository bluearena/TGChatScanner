package clarifaiAPI

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const APIURL = "https://api.clarifai.com/v2/models/aaa03c23b3724a16a56b629203edc62c/outputs"

type ClarifaiAPI struct {
	httpClient *http.Client
	APIKey     string
}

func NewClarifaiAPI(APIKey string) *ClarifaiAPI {
	return &ClarifaiAPI{
		httpClient: &http.Client{},
		APIKey:     APIKey,
	}
}

func (ca *ClarifaiAPI) RecognizeImage(URL string, minProbability float64) ([]string, error) {
	result, err := ca.RecognizeImages([]string{URL}, minProbability)

	if err != nil {
		return nil, err
	}

	return result[0], nil
}

// Recognize images and return tags
func (ca *ClarifaiAPI) RecognizeImages(URLs []string, minProbability float64) ([][]string, error) {
	requestBody := Request{make([]Input, len(URLs))}

	for i, URL := range URLs {
		requestBody.Inputs[i] = Input{Data: Data{Image{URL}}}
	}

	b, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", APIURL, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Key "+ca.APIKey)
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
			err := errors.New(output.Input.Data.Image.URL + ": " + output.Status.Description)
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
