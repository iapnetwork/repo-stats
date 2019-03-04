package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// setRequest creates and sends the request.
func setRequest(uri string, token string) []byte {
	// Set up the request.
	client := &http.Client{}
	request, err := http.NewRequest("GET", uri, nil)
	check(err)
	request.Header.Set("Authorization", token)

	// Get the repos json data.
	response, _ := client.Do(request)
	jsonData, _ := ioutil.ReadAll(response.Body)

	return jsonData
}

// getJsonResponse gets the full json object returned from the Request.
func getJsonResponse(uri string, token string, fixer string) []interface{} {
	// Set up the request.
	jsonData := setRequest(uri, token)
	jsonText := string(jsonData)

	// Github returns empty stats data for the first uncached request, so try again.
	if jsonText == "{}" {
		jsonData = setRequest(uri, token)
		jsonText = string(jsonData)
	}

	// Fix the bad json from GitHub.
	jsonText = `{ "` + fixer + `": ` + jsonText + ` }`

	// Unmarshal the json.
	var jsonDataFixed map[string]interface{}
	err := json.Unmarshal([]byte(jsonText), &jsonDataFixed)
	check(err)

	// Use type assertion to get the repo list into the correct type.
	dataList := jsonDataFixed[fixer].([]interface{})
	return dataList
}
