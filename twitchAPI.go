package main

import "bytes"
import "net/http"
import "strconv"
import "errors"
import "io/ioutil"
import "encoding/json"
import "fmt"

const TWITCH_USER_INFO_URL = "https://api.twitch.tv/helix/users?id="
const DISPLAY_NAME = "display_name"

func getUserInformation(requestId string, userId int) (userInformation, error) {
	userInfo := userInformation{}
	var byteBuffer bytes.Buffer
	byteBuffer.WriteString(TWITCH_USER_INFO_URL)
	byteBuffer.WriteString(strconv.Itoa(userId))
	url := byteBuffer.String()

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return userInfo, err
	}

	byteBuffer.Reset()
	byteBuffer.WriteString("Bearer ")
	byteBuffer.WriteString(config.Oauth)
	request.Header.Set("Authorization", byteBuffer.String())

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return userInfo, err
	}

	if response.StatusCode != 200 {
		byteBuffer.Reset()
		byteBuffer.WriteString("Unsuccessful response. Got ")
		byteBuffer.WriteString(response.Status)
		return userInfo, errors.New(byteBuffer.String())
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return userInfo, err
	}

	fmt.Println(requestId, "getUserInformation response Status:", response.Status)
	fmt.Println(requestId, "getUserInformation response Headers:", response.Header)
	fmt.Println(requestId, "getUserInformation response Body:", string(responseBody))

	var data map[string]interface{}
	if err := json.Unmarshal(responseBody, &data); err != nil {
		return userInfo, err
	}

	dataProperty := data["data"].([]interface{})
	if len(dataProperty) == 0 {
		return userInfo, errors.New("Json body is not formatted as expected.")
	}

	dataStruct := dataProperty[0].(map[string]interface{})
	if displayName, ok := dataStruct[DISPLAY_NAME]; ok {
		userInfo.displayName = displayName.(string)
		return userInfo, nil
	}

	return userInfo, errors.New("Could not find display name in twitch lookup")
}
