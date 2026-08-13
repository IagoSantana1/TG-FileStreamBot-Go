package commands

import (
	"EverythingSuckz/fsb/internal/utils"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
)

func processBatch(
	chatID int64,
	items []BatchItem,
	loadingCtx *ext.Context,
	loadingMsgID int,
) {

	if len(items) == 0 {
		return
	}

	media := make([]utils.MediaCopyItem, 0, len(items))

	for _, item := range items {

		media = append(media, utils.MediaCopyItem{
			Media:   item.Update.EffectiveMessage.Media,
			Caption: item.Update.EffectiveMessage.Message.Message,
		})

	}

	copied, err := utils.SendMediaCopy(items[0].Ctx, chatID, media)

	if err != nil {
		utils.Logger.Sugar().Error(err)
		return
	}

	if len(copied) != len(items) {
		utils.Logger.Sugar().Errorf(
			"Quantidade de mídias retornadas (%d) diferente da enviada (%d)",
			len(copied),
			len(items),
		)
		return
	}

	// Remove a mensagem "Agrupando arquivos..."
	if loadingCtx != nil && loadingMsgID != 0 {

		_, err := loadingCtx.Raw.MessagesDeleteMessages(
			loadingCtx,
			&tg.MessagesDeleteMessagesRequest{
				ID:     []int{loadingMsgID},
				Revoke: true,
			},
		)

		if err != nil {
			utils.Logger.Sugar().Error(err)
		}
	}

	// Responde o usuário
	for i := range items {

		if err := processCopiedMedia(
			items[i],
			copied[i],
		); err != nil {

			utils.Logger.Sugar().Error(err)

		}
	}
}
