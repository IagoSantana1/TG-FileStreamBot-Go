package commands

import (
	"EverythingSuckz/fsb/internal/utils"
)

func processBatch(chatID int64, items []BatchItem) {

    media := make([]utils.MediaCopyItem, 0, len(items))

    for _, item := range items {

        media = append(media, utils.MediaCopyItem{
            Media: item.Update.EffectiveMessage.Media,
            Caption: item.Update.EffectiveMessage.Message.Message,
        })

    }

    updates, err := utils.SendMediaCopy(
        items[0].Ctx,
        chatID,
        media,
    )

    if err != nil {
        return
    }

    copied, err := utils.ParseMediaUpdates(updates)
    if err != nil {
        return
    }

    if len(copied) != len(items) {
        return
    }

    for i := range items {

        processCopiedMedia(
            items[i],
            copied[i],
        )

    }

}
