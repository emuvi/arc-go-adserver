package guide

import (
	"os"
	"strconv"
)

type aCommand struct {
	name         string
	isFlag       bool
	shortKey     string
	longKey      string
	description  string
	defaultValue string
	value        string
}

func Command(name string, isFlag bool, shortKey, longKey string) *aCommand {
	return &aCommand{
		name:         name,
		isFlag:       isFlag,
		shortKey:     shortKey,
		longKey:      longKey,
		description:  "",
		defaultValue: "",
		value:        "",
	}
}

func (command *aCommand) Description(description string) *aCommand {
	command.description = description
	return command
}

func (command *aCommand) DefaultValue(defaultValue string) *aCommand {
	command.defaultValue = defaultValue
	return command
}

type aCommands struct {
	data       []*aCommand
	freeParams []string
}

var Commands = &aCommands{
	data:       []*aCommand{},
	freeParams: nil,
}

func (commands *aCommands) Add(command *aCommand) *aCommands {
	commands.data = append(commands.data, command)
	return commands
}

func (commands *aCommands) Parse() *aCommands {
	for _, command := range commands.data {
		if command.isFlag {
			command.defaultValue = "false"
		}
		command.value = ""
	}
	freeParams := []string{}
	for i := 0; i < len(os.Args); i++ {
		arg := os.Args[i]
		found := false
		for _, command := range commands.data {
			if command.shortKey == arg || command.longKey == arg {
				if command.isFlag {
					command.value = "true"
				} else if i < len(os.Args)-1 {
					i++
					command.value = os.Args[i]
				}
				found = true
				break
			}
		}
		if !found {
			freeParams = append(freeParams, arg)
		}
	}
	commands.freeParams = freeParams
	return commands
}

func (commands *aCommands) PutOnConfigs() *aCommands {
	return commands.Put(Configs)
}

func (commands *aCommands) Put(onConfigs *aConfigs) *aCommands {
	for _, command := range commands.data {
		if command.value != "" {
			onConfigs.SetString(command.name, command.value)
		} else {
			onConfigs.SetString(command.name, command.defaultValue)
		}
	}
	for index, freeParam := range commands.freeParams {
		onConfigs.SetString("CommmandsFreeParam"+strconv.Itoa(index), freeParam)
	}
	onConfigs.SetInt("CommmandsFreeParamsSize", len(commands.freeParams))
	return commands
}

func GetCommandsFreeParamsSize(ofConfigs *aConfigs) int {
	return ofConfigs.GetInt("CommmandsFreeParamsSize", 0)
}

func GetCommandsFreeParam(ofConfigs *aConfigs, index int) string {
	return ofConfigs.GetString("CommmandsFreeParam"+strconv.Itoa(index), "")
}
