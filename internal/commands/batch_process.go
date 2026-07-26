package commands

import (
	"EverythingSuckz/fsb/internal/utils"
	"fmt"

	"github.com/gotd/td/tg"
)

type CopiedMedia struct {
	MessageID int
	Message   *tg.Message
}

func processBatch(chatID int64, items []BatchItem) {

	if len(items) == 0 {
		return
	}

	fmt.Printf("Processando lote com %d arquivos\n", len(items))

	mediaItems := make([]utils.MediaCopyItem, 0, len(items))

	for _, item := range items {
		msg := item.Update.EffectiveMessage

		mediaItems = append(mediaItems, utils.MediaCopyItem{
			Media:   msg.Media,
			Caption: msg.Message.Message,
		})

	}

	updates, err := utils.SendMediaCopy(items[0].Ctx, chatID, mediaItems)

	if err != nil {
		utils.Logger.Sugar().Error(err)
		return
	}

	fmt.Printf("Telegram retornou %d updates\n", len(updates.Updates))
}

func ParseMediaUpdates(updates *tg.Updates) ([]CopiedMedia, error)
