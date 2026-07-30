package commands

import (
	"fmt"
	"strings"

	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

func processCopiedMedia(
	original BatchItem,
	copied utils.CopiedMedia,
) error {

	file, err := utils.FileFromMedia(
		copied.Message.Media,
		copied.Message.Message,
	)
	if err != nil {
		return err
	}

	fullHash := utils.PackFile(
		file.FileName,
		file.FileSize,
		file.MimeType,
		file.ID,
	)

	metadata := utils.DetectFileMetadata(
		file.FileName,
		copied.Message.Message,
	)

	displayName := utils.FormatFileNameForDisplay(metadata)
	if displayName == "" {
		displayName = file.FileName
	}

	hash := utils.GetShortHash(fullHash)
	strmFileName := utils.ProcessStrmFileName(displayName)
	strmFileNameWithExt := strmFileName + ".strm"
	linkStrm := buildStrmLink(copied.MessageID, hash, strmFileNameWithExt)
	link := fmt.Sprintf("%s/stream/%d?hash=%s", config.ValueOf.Host, copied.MessageID, hash)

	messageFormatted := []styling.StyledTextOption{
		styling.Bold("🎬 Mídia Pronta para Acesso"),
		styling.Plain("\n➖➖➖➖➖➖➖➖➖➖➖\n"),
		styling.Bold("📁 Arquivo: "),
		styling.Code(file.FileName),
		styling.Plain("\n\n"),
		styling.Bold("Nome do STRM: "),
		styling.Code(strmFileNameWithExt),
		styling.Plain("\n\n➖➖➖➖➖➖➖➖➖➖➖\n"),
		styling.Bold("🔗 Links Rápidos (Toque para copiar):\n\n"),
		styling.Bold("📺 Stream: "),
		styling.Code(link),
		styling.Plain("\n\n"),
		styling.Bold("⬇️ Download: "),
		styling.Code(link + "&d=true"),
	}

	text := styling.Code(link)

	row := tg.KeyboardButtonRow{
		Buttons: []tg.KeyboardButtonClass{
			&tg.KeyboardButtonURL{
				Text: "Download",
				URL:  link + "&d=true",
			},
			&tg.KeyboardButtonURL{
				Text: "STRM",
				URL:  linkStrm,
			},
		},
	}

	if strings.Contains(file.MimeType, "video") ||
		strings.Contains(file.MimeType, "audio") ||
		strings.Contains(file.MimeType, "pdf") {

		row.Buttons = append(row.Buttons,
			&tg.KeyboardButtonURL{
				Text: "Stream",
				URL:  link,
			},
		)
	}

	markup := &tg.ReplyInlineMarkup{
		Rows: []tg.KeyboardButtonRow{
			row,
		},
	}

	if strings.Contains(link, "http://localhost") {

		_, err = original.Ctx.Reply(
			original.Update,
			ext.ReplyTextStyledText(text),
			&ext.ReplyOpts{
				NoWebpage:        false,
				ReplyToMessageId: original.Update.EffectiveMessage.ID,
			},
		)

	} else {

		_, err = original.Ctx.Reply(
			original.Update,
			ext.ReplyTextStyledTextArray(messageFormatted),
			&ext.ReplyOpts{
				Markup:           markup,
				NoWebpage:        false,
				ReplyToMessageId: original.Update.EffectiveMessage.ID,
			},
		)

	}

	if err != nil {

		utils.Logger.Sugar().Error(err)

		_, _ = original.Ctx.Reply(
			original.Update,
			ext.ReplyTextString(fmt.Sprintf("Error - %s", err.Error())),
			nil,
		)

		return err
	}

	return nil
}
