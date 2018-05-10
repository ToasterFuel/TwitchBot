package main

import "github.com/gorilla/websocket"
import "bytes"
import "strings"
import "fmt"
import "errors"
import "time"

var toTwitchChannel chan string
var fromTwitchChannel chan string
var shouldListen bool
var commands botCommands

const STOP_IRC = "STOP_IRC"

type messageDetails struct {
	userName    string
	message     string
	channelName string
}

func stopIRC() {
	toTwitchChannel <- STOP_IRC
}

func sendThankYou(toUserId string) {
	sendPrivateMessage(toUserId, "Hey! Thanks for the follow! I really appreciate your support :)")
}

func sendChannelMessage(message string) {
	toTwitchChannel <- string(getMessageCommand(message))
}

func sendPrivateMessage(toUserId string, message string) {
	toTwitchChannel <- string(getDirectMessageCommand(toUserId, message))
}

func createChatConnection(requestId string) error {
	if toTwitchChannel != nil || fromTwitchChannel != nil {
		fmt.Println(requestId, "Twitch chat already created. Nothing to do.")
		return nil
	}

	var err error
	if commands, err = readBotCommands(); err != nil {
		return err
	}
	connection, response, err := websocket.DefaultDialer.Dial("wss://irc-ws.chat.twitch.tv:443", nil)

	if err != nil {
		fmt.Println(requestId, "Could not connect to Twitch IRC endpoint", err)
		fmt.Println(requestId, "Response from Twitch IRC endpoint", response)
	}

	if err = initConnection(requestId, connection); err != nil {
		connection.Close()
		return err
	}

	shouldListen = true
	toTwitchChannel = make(chan string, 100)
	fromTwitchChannel = make(chan string, 100)
	go listenForMessages(connection)
	go processToMessages(connection)
	go processFromMessages()

	sendPrivateMessage(config.ChannelName, "I have started the IRC bot")

	return nil
}

func initConnection(requestId string, connection *websocket.Conn) error {
	if err := connection.WriteMessage(websocket.TextMessage, getPassCommand()); err != nil {
		return err
	}
	if err := connection.WriteMessage(websocket.TextMessage, getNickCommand()); err != nil {
		return err
	}
	if err := readAndPrintMessages(requestId, connection); err != nil {
		return err
	}
	if err := connection.WriteMessage(websocket.TextMessage, getJoinCommand()); err != nil {
		return err
	}

	return nil
}

func readAndPrintMessages(requestId string, connection *websocket.Conn) error {
	_, message, err := connection.ReadMessage()
	if err != nil {
		return err
	}
	fmt.Println(requestId, "Twitch message received: ", string(message[:len(message)]))
	return nil
}

func listenForMessages(connection *websocket.Conn) {
	for shouldListen {
		_, message, err := connection.ReadMessage()
		if err != nil {
			closeConnections(connection, err)
			return
		}
		fromTwitchChannel <- string(message[:len(message)])
	}
	closeConnections(connection, nil)
}

func processFromMessages() {
	var byteBuffer bytes.Buffer
	for item := range fromTwitchChannel {
		byteBuffer.Reset()
		fmt.Println("ReceivedFromTwitch:", item)
		if strings.Contains(item, "PING") {
			toTwitchChannel <- string(getCommand("PONG :tmi.twitch.tv"))
			continue
		}

		messageDeets, err := parseMessage(item)
		if err != nil {
			fmt.Println("Error parsing message from Twitch: ", err)
		}
		lowerMessage := strings.ToLower(messageDeets.message)
		if strings.Contains(lowerMessage, "!") {
			processSpecificCommand(messageDeets)
		}
	}
}

func processToMessages(connection *websocket.Conn) {
	for item := range toTwitchChannel {
		if item == STOP_IRC {
			lastMessage := getDirectMessageCommand(config.ChannelName, "I am shutting down the IRC bot")
			connection.WriteMessage(websocket.TextMessage, lastMessage)
			time.Sleep(10 * time.Second)
			closeConnections(connection, errors.New("Manually stopping!"))
			break
		}
		fmt.Println("WritingToTwitch:", item)
		if err := connection.WriteMessage(websocket.TextMessage, []byte(item)); err != nil {
			closeConnections(connection, err)
		}
	}
}

func processSpecificCommand(messageDeets messageDetails) {
	splitStrings := strings.Split(messageDeets.message, "!")
	if len(splitStrings) == 0 {
		fmt.Println("Not a command!!!")
		return
	}
	possibleCommand := strings.TrimSpace(splitStrings[1])

	if value, ok := commands.CommandMap[possibleCommand]; ok {
		toTwitchChannel <- string(getMessageCommand(value))
	} else {
		fmt.Println("Unknown command: ", possibleCommand)
	}
}

func parseMessage(message string) (messageDetails, error) {
	messageDeets := messageDetails{}
	//ReceivedFromTwitch: :<username>!<username>@<username>.tmi.twitch.tv PRIVMSG #<channel> :<message>
	if !strings.Contains(message, "PRIVMSG") {
		return messageDeets, errors.New("Unknown message type.")
	}
	splitStrings := strings.Split(message, "!")
	if len(splitStrings) == 0 {
		return messageDeets, errors.New("Unknown message format 1.")
	}
	splitStrings = strings.Split(splitStrings[0], ":")
	if len(splitStrings) != 2 {
		return messageDeets, errors.New("Unknown message format 2.")
	}
	messageDeets.userName = splitStrings[1]

	splitStrings = strings.Split(message, "#")
	if len(splitStrings) != 2 {
		return messageDeets, errors.New("Unknown message format 3.")
	}
	channelAndMessage := splitStrings[1]
	splitStrings = strings.Split(channelAndMessage, " ")
	if len(splitStrings) == 0 {
		return messageDeets, errors.New("Unknown message format 4.")
	}
	messageDeets.channelName = splitStrings[0]

	splitStrings = strings.Split(channelAndMessage, ":")
	if len(splitStrings) != 2 {
		return messageDeets, errors.New("Unknown message format 5.")
	}
	messageDeets.message = splitStrings[1]

	return messageDeets, nil
}

func closeConnections(connection *websocket.Conn, err error) {
	fmt.Println("Could not read from Twitch connection. Closing connection.", err)
	if toTwitchChannel != nil {
		close(toTwitchChannel)
	}
	if fromTwitchChannel != nil {
		close(fromTwitchChannel)
	}
	toTwitchChannel = nil
	fromTwitchChannel = nil
	connection.Close()
}

func getPassCommand() []byte {
	return getCommand("PASS oauth:", config.Oauth)
}

func getNickCommand() []byte {
	return getCommand("NICK ", config.UserName)
}

func getJoinCommand() []byte {
	return getCommand("JOIN #", config.ChannelName)
}

func getDirectMessageCommand(toUser string, message string) []byte {
	return getCommand("PRIVMSG #jtv :/w ", toUser, " ", message)
}

func getMessageCommand(message string) []byte {
	return getCommand("PRIVMSG #", config.ChannelName, " :", message)
}

func getCommand(strs ...string) []byte {
	allStrings := make([]string, len(strs)+1)
	for _, value := range strs {
		allStrings = append(allStrings, value)
	}
	allStrings = append(allStrings, "\r\n")
	return []byte(strings.Join(allStrings, ""))
}
