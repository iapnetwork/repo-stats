package main

import (
	"encoding/json"
	"io/ioutil"
)

// Configuration contains the config properties.
type Configuration struct {
	URIRepos string `json:"uri_repos"`
	URIStats string `json:"uri_stats"`
	Token    string `json:"token"`
}

// setConfiguration adds all of the json config settings into the struct.
func setConfiguration() Configuration {
	// Declare the configuration struct,
	var configuration Configuration

	// Set the data into the struct.
	configData := getConfigData("config.json")
	configuration.URIRepos = configData["uri_repos"].(string)
	configuration.URIStats = configData["uri_stats"].(string)

	secretData := getConfigData("secret.json")
	configuration.Token = "Bearer " + secretData["token"].(string)
	return configuration
}

func getConfigData(fileName string) map[string]interface{} {
	// Get the config json into the Configuration struct.
	jsonData := getConfiguration(fileName)
	var configData map[string]interface{}
	err := json.Unmarshal(jsonData, &configData)
	check(err)
	return configData
}

func getConfiguration(fileName string) []byte {
	// Read the contents of the config json.
	byteValue, err := ioutil.ReadFile(fileName)
	check(err)
	return byteValue
}
