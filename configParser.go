package main

import "encoding/json"
import "io/ioutil"

type configInformation struct {
	Oauth          string `json:"Oauth"`
	UserName       string `json:"UserName"`
	ChannelName    string `json:"ChannelName"`
	CallbackUserId string `json:"CallbackUserId"`
	CallbackUrl    string `json:"CallbackUrl"`
	Secret         string `json:"Secret"`
}

func readConfigInformation() (configInformation, error) {
	return readConfigByFile("config.json")
}

func readConfigByFile(filePath string) (configInformation, error) {
	info := configInformation{}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return info, err
	}

	if err = json.Unmarshal(data, &info); err != nil {
		return info, err
	}

	return info, nil
}
