package main

import "encoding/json"
import "fmt"
import "errors"

type twitchNotificationType uint8

const USER_ID = "user_id"
const FROM_ID = "from_id"
const TO_ID = "to_id"
const FOLLOWED_AT = "followed_at"

const (
	FOLLOW twitchNotificationType = iota
	STREAM_UP
	STREAM_DOWN
	UNKNOWN
)

type userInformation struct {
	displayName string
}

type twitchNotification struct {
	notificationType twitchNotificationType
	data             map[string]string
}

func getNotification(requestId string, body []byte) (twitchNotification, error) {
	var data map[string]interface{}
	notification := twitchNotification{UNKNOWN, make(map[string]string)}

	if err := json.Unmarshal(body, &data); err != nil {
		return notification, err
	}

	fmt.Println(requestId, "all data:", data)
	dataProperty := data["data"].([]interface{})
	if len(dataProperty) == 0 {
		notification.notificationType = STREAM_DOWN
		return notification, nil
	}

	dataStruct := dataProperty[0].(map[string]interface{})
	if userId, ok := dataStruct[USER_ID]; ok {
		notification.notificationType = STREAM_UP
		notification.data[USER_ID] = userId.(string)
		return notification, nil
	}

	if fromId, ok := dataStruct[FROM_ID]; ok {
		if toId, ok := dataStruct[TO_ID]; ok {
			if followedAt, ok := dataStruct[FOLLOWED_AT]; ok {
				notification.notificationType = FOLLOW
				notification.data[FROM_ID] = fromId.(string)
				notification.data[TO_ID] = toId.(string)
				notification.data[FOLLOWED_AT] = followedAt.(string)
				return notification, nil
			}
		}
	}

	return notification, errors.New("Unknown notification type")
}
