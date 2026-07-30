package utils

import (
	"fmt"

	"github.com/gotd/td/tg"
)

type CopiedMedia struct {
	MessageID int
	Message   *tg.Message
}

func ParseMediaUpdates(updates *tg.Updates) ([]CopiedMedia, error) {

	var messages []*tg.Message

	for _, update := range updates.Updates {

		switch u := update.(type) {

		case *tg.UpdateNewMessage:

			if msg, ok := u.Message.(*tg.Message); ok {
				messages = append(messages, msg)
			}

		case *tg.UpdateNewChannelMessage:

			if msg, ok := u.Message.(*tg.Message); ok {
				messages = append(messages, msg)
			}

		}
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("nenhuma mensagem encontrada nos updates")
	}

	result := make([]CopiedMedia, 0, len(messages))

	for _, msg := range messages {

		result = append(result, CopiedMedia{
			MessageID: msg.ID,
			Message:   msg,
		})

	}

	return result, nil
}
