package commands

import (
	"fmt"

	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

// minItemsForZipButton define a partir de quantos arquivos processados o
// botão de "baixar tudo em .zip" passa a ser oferecido. Para um único
// arquivo, o botão de STRM individual já resolve.
const minItemsForZipButton = 2

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

	// Responde o usuário e coleta as entradas .strm para o zip do lote
	strmEntries := make([]utils.BatchStrmEntry, 0, len(items))
	titleCounts := make(map[string]int)

	for i := range items {

		entry, title, err := processCopiedMedia(
			items[i],
			copied[i],
		)

		if err != nil {
			utils.Logger.Sugar().Error(err)
			continue
		}

		if entry != nil {
			strmEntries = append(strmEntries, *entry)
		}

		if title != "" {
			titleCounts[title]++
		}
	}

	seriesName := mostFrequentTitle(titleCounts)

	sendBatchZipButton(items[0].Ctx, chatID, strmEntries, seriesName)
}

// mostFrequentTitle retorna o título mais comum entre os arquivos do lote
// (ex: quando todos os episódios de uma série são detectados com o mesmo
// título). Em caso de empate, retorna qualquer um dos mais frequentes.
func mostFrequentTitle(counts map[string]int) string {
	best := ""
	bestCount := 0

	for title, count := range counts {
		if count > bestCount {
			best = title
			bestCount = count
		}
	}

	return best
}

// sendBatchZipButton envia uma mensagem final resumindo o lote processado,
// com um botão que baixa um .zip contendo todos os arquivos .strm gerados.
// seriesName, quando detectado, é usado para nomear o arquivo .zip.
func sendBatchZipButton(ctx *ext.Context, chatID int64, entries []utils.BatchStrmEntry, seriesName string) {

	if len(entries) < minItemsForZipButton {
		return
	}

	token, err := utils.RegisterBatchStrm(entries, seriesName)
	if err != nil {
		utils.Logger.Sugar().Error(err)
		return
	}

	zipLink := fmt.Sprintf("%s/batch-strm/%s", config.ValueOf.Host, token)

	summary := fmt.Sprintf("✅ %d arquivos processados", len(entries))
	if seriesName != "" {
		summary = fmt.Sprintf("✅ %d arquivos processados — %s", len(entries), seriesName)
	}

	messageFormatted := []styling.StyledTextOption{
		styling.Bold("📦 Lote processado!"),
		styling.Plain("\n\n"),
		styling.Bold(summary),
		styling.Plain("\n\n"),
		styling.Bold("💾 baixe todos os arquivos .strm de uma vez:"),
	}

	// SendMessage trabalha com o request "cru" da API (texto + entities),
	// diferente de ctx.Reply, que aceita []styling.StyledTextOption
	// diretamente. Por isso convertemos o texto estilizado manualmente
	// aqui com entity.Builder antes de montar o request.
	tb := entity.Builder{}
	if err := styling.Perform(&tb, messageFormatted...); err != nil {
		utils.Logger.Sugar().Error(err)
		return
	}
	text, entities := tb.Complete()

	_, err = ctx.SendMessage(chatID, &tg.MessagesSendMessageRequest{
		Message:  text,
		Entities: entities,
		ReplyMarkup: &tg.ReplyInlineMarkup{
			Rows: []tg.KeyboardButtonRow{
				{
					Buttons: []tg.KeyboardButtonClass{
						&tg.KeyboardButtonURL{
							Text: "📦 Baixar .zip",
							URL:  zipLink,
						},
					},
				},
			},
		},
	})

	if err != nil {
		utils.Logger.Sugar().Error(err)
	}
}
