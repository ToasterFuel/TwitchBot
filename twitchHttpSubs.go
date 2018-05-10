package main

import "fmt"
import "encoding/json"
import "net/http"
import "bytes"
import "strconv"
import "io/ioutil"
import "errors"

const TWITCH_WEBHOOK_URL = "https://api.twitch.tv/helix/webhooks/hub"
const SUBSCRIBE = "subscribe"
const UNSUBSCRIBE = "unsubscribe"
const FOLLOW_TOPIC = "https://api.twitch.tv/helix/users/follows?first=1&to_id="
const STREAM_UP_DOWN_TOPIC = "https://api.twitch.tv/helix/streams?user_id="

type subscribeBody struct {
	Mode         string `json:"hub.mode"`
	Topic        string `json:"hub.topic"`
	Callback     string `json:"hub.callback"`
	LeaseSeconds string `json:"hub.lease_seconds"`
	Secret       string `json:"hub.secret"`
}

func subscribeToFollowerNotifications(requestId string, leaseSeconds int) error {
	var byteBuffer bytes.Buffer
	byteBuffer.WriteString(FOLLOW_TOPIC)
	byteBuffer.WriteString(config.CallbackUserId)
	topic := byteBuffer.String()

	return subscribe(requestId, topic, leaseSeconds)
}

func subscribeToScreamUpDownNotifications(requestId string, leaseSeconds int) error {
	var byteBuffer bytes.Buffer
	byteBuffer.WriteString(STREAM_UP_DOWN_TOPIC)
	byteBuffer.WriteString(config.CallbackUserId)
	topic := byteBuffer.String()

	return subscribe(requestId, topic, leaseSeconds)
}

func subscribe(requestId string, topic string, leaseSeconds int) error {
	var byteBuffer bytes.Buffer
	body := subscribeBody{}
	body.Mode = SUBSCRIBE
	body.Callback = config.CallbackUrl
	body.LeaseSeconds = strconv.Itoa(leaseSeconds)
	body.Secret = config.Secret
	body.Topic = topic

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", TWITCH_WEBHOOK_URL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	byteBuffer.WriteString("Bearer ")
	byteBuffer.WriteString(config.Oauth)
	request.Header.Set("Authorization", byteBuffer.String())
	request.Header.Set("Content-Type", "application/json")
	byteBuffer.Reset()

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	fmt.Println(requestId, "subscribeToFollowerNotification response Status:", response.Status)
	fmt.Println(requestId, "subscribeToFollowerNotification response Headers:", response.Header)
	fmt.Println(requestId, "subscribeToFollowerNotification response Body:", string(responseBody))

	if response.StatusCode != 202 {
		byteBuffer.Reset()
		byteBuffer.WriteString("Unsuccessful return code while subscribing. Got ")
		byteBuffer.WriteString(response.Status)
		return errors.New(byteBuffer.String())
	}

	return nil
}
