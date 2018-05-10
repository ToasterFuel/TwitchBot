package main

import "encoding/json"
import "io/ioutil"

type botCommands struct {
	CommandMap map[string]string `json:"CommandMap"`
}

func readBotCommands() (botCommands, error) {
	return readBotCommandsByFile("botCommands.json")
}

func readBotCommandsByFile(filePath string) (botCommands, error) {
	commands := botCommands{}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return commands, err
	}

	if err = json.Unmarshal(data, &commands); err != nil {
		return commands, err
	}

	return commands, nil
}
