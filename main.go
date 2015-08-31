package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/docopt/docopt-go"
	"github.com/nlopes/slack"
)

const usage = `Slack API Interface.
	
Sets topic for specified channel.
	
Usage:
	$0 -h | --help
	$0 [options] -k <token> -C <channel> [-t=]

Options:
    -h --help     Show this help.
    -C            Channel operations.
      -t=<topic>  Sets topic for channel.
	              Supports -i flag and templating capability.
    -k=<token>    Slack API Token.
    -i            Read stdin for additional parameters encoded in JSON.
	              Useful for setting template topic names, like
				  with flag -t 'Man on duty: {{.name}}'.
`

type API struct {
	*slack.Client

	additionalParameters map[string]interface{}
}

func main() {
	args, err := docopt.Parse(
		strings.Replace(usage, "$0", os.Args[0], -1),
		nil, true, "slack-topic-setter 1.0", false,
	)
	if err != nil {
		panic(err)
	}

	var additionalParameters map[string]interface{}
	if args["-i"].(bool) {
		stdinDecoder := json.NewDecoder(os.Stdin)
		err := stdinDecoder.Decode(&additionalParameters)
		if err != nil {
			log.Printf("error reading stdin: %s", err)
			return
		}
	}

	api := &API{
		slack.New(args["-k"].(string)),
		additionalParameters,
	}

	switch {
	case args["-C"]:
		err = api.handleChannelMode(args)
	}

	if err != nil {
		log.Printf("error: %s", err)
	}
}

func (api *API) handleChannelMode(args map[string]interface{}) error {
	switch {
	case args["-t"] != nil:
		return api.setChannelTopic(
			args["<channel>"].(string),
			args["-t"].(string),
		)

	}

	return nil
}

func (api *API) setChannelTopic(channelName, topic string) error {
	channels, err := api.GetChannels(true)
	if err != nil {
		return err
	}

	if api.additionalParameters != nil {
		topicTemplate, err := template.New("topic").Parse(topic)
		if err != nil {
			return err
		}

		buffer := &bytes.Buffer{}
		err = topicTemplate.Execute(buffer, api.additionalParameters)
		if err != nil {
			return err
		}

		topic = buffer.String()
	}

	for _, channel := range channels {
		if channel.Name == channelName {
			api.SetChannelTopic(channel.ID, topic)
			return nil
		}
	}

	return fmt.Errorf("channel not found: %s", channelName)
}
