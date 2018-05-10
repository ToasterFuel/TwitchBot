package main

import "net/http"
import "io/ioutil"
import "fmt"
import "strings"
import "strconv"

const CHALLENGE = "hub.challenge"
const SIGNATURE_HEADER = "X-Hub-Signature"

func handleCallback(responseWriter http.ResponseWriter, request *http.Request) {
	requestId := getRequestId()
	body, err := ioutil.ReadAll(request.Body)
	headers := request.Header
	queryParameters := request.URL.Query()
	fmt.Println("\nNew request", requestId)
	fmt.Println(requestId, "headers:", headers)
	fmt.Println(requestId, "url query parameters:", queryParameters)
	if err != nil {
		fmt.Println(requestId, "Error parsing body.", err)
		responseWriter.WriteHeader(400)
		return
	}
	fmt.Println(requestId, "body:", string(body[:len(body)]))

	challengeArray := queryParameters[CHALLENGE]
	if len(challengeArray) > 0 {
		handleConfirmation(requestId, responseWriter, challengeArray[0])
		return
	}
	handleNotification(requestId, responseWriter, body, headers)
}

func handleConfirmation(requestId string, responseWriter http.ResponseWriter, challenge string) {
	fmt.Println(requestId, "Writing the challenge back to confirmation")
	responseWriter.Write([]byte(challenge))
	responseWriter.WriteHeader(200)
}

func handleNotification(requestId string, responseWriter http.ResponseWriter, body []byte, headers http.Header) {
	signatureMatch := false
	signatureValue := headers[SIGNATURE_HEADER]
	if len(signatureValue) > 0 {
		split := strings.Split(signatureValue[0], "=")
		if len(split) == 2 {
			signatureMatch = checkSignature(body, split[1], []byte(config.Secret))
		}
	} else {
		fmt.Println(requestId, "Length of signature is 0")
	}
	if !signatureMatch {
		fmt.Println(requestId, "Signature does not match. Ignoring request")
		return
	}

	notification, err := getNotification(requestId, body)
	if err != nil || notification.notificationType == UNKNOWN {
		fmt.Println(requestId, "Could not determine the type of notification.", err)
		responseWriter.WriteHeader(400)
		return
	}

	switch notification.notificationType {
	case FOLLOW:
		if err = handleFollow(requestId, notification); err != nil {
			fmt.Println(requestId, "Error prossing follow notification", err)
			responseWriter.WriteHeader(400)
			return
		}
	case STREAM_UP:
		if err = handleStreamUp(requestId, notification); err != nil {
			fmt.Println(requestId, "Error prossing stream upnotification", err)
			responseWriter.WriteHeader(400)
			return
		}
	case STREAM_DOWN:
		if err = handleStreamDown(requestId, notification); err != nil {
			fmt.Println(requestId, "Error prossing stream down notification", err)
			responseWriter.WriteHeader(400)
			return
		}
	default:
		fmt.Println(requestId, "Could not determine the try of notification.", err)
		responseWriter.WriteHeader(400)
		return
	}

	responseWriter.WriteHeader(200)
}

func handleFollow(requestId string, notification twitchNotification) error {
	fromId, err := strconv.Atoi(notification.data[FROM_ID])
	if err != nil {
		fmt.Println(requestId, "Could not convert fromId to int.", notification.data[FROM_ID])
		return err
	}
	userInfo, err := getUserInformation(requestId, fromId)
	if err != nil {
		return err
	}
	sendThankYou(userInfo.displayName)
	return nil
}

func handleStreamUp(requestId string, notification twitchNotification) error {
	return createChatConnection(requestId)
}

func handleStreamDown(requestId string, notification twitchNotification) error {
	stopIRC()
	return nil
}
