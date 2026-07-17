package commands

import (
	"fmt"
	"net/url"
	"strings"

	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/celestix/gotgproto/types"

	// "github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

func (m *command) LoadStream(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(
		handlers.NewMessage(nil, sendLink),
	)
}

func supportedMediaFilter(m *types.Message) (bool, error) {
	if not := m.Media == nil; not {
		return false, dispatcher.EndGroups
	}
	switch m.Media.(type) {
	case *tg.MessageMediaDocument:
		return true, nil
	case *tg.MessageMediaPhoto:
		return true, nil
	case tg.MessageMediaClass:
		return false, dispatcher.EndGroups
	default:
		return false, nil
	}
}

// função para criar o link do arquivo strm para download
func buildStrmLink(messageID int, hash string, strmFileName string) string {
	name := strings.TrimSpace(strmFileName)
	if name == "" {
		return ""
	}

	if !strings.HasSuffix(strings.ToLower(name), ".strm") {
		name += ".strm"
	}

	encodedName := url.QueryEscape(name)

	return fmt.Sprintf("%s/strm/%d?hash=%s&name=%s", config.ValueOf.Host, messageID, hash, encodedName)
}

// função para enviar o link do video no chat do telegram
func sendLink(ctx *ext.Context, u *ext.Update) error {

	// Captura e identifica o ID do chat e do usuário
	chatId := u.EffectiveChat().GetID()
	peerChatId := ctx.PeerStorage.GetPeerById(chatId)

	// Garante que o bot só responda em conversas privadas (Direct Messages)
	// e nao responda no canal de log
	if peerChatId.Type != int(storage.TypeUser) {
		return dispatcher.EndGroups
	}

	if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
		ctx.Reply(u, ext.ReplyTextString("You are not allowed to use this bot."), nil)
		return dispatcher.EndGroups
	}
	supported, err := supportedMediaFilter(u.EffectiveMessage)
	if err != nil {
		return err
	}
	if !supported {
		ctx.Reply(u, ext.ReplyTextString("Desculpe, este tipo de mensagem não é suportado."), nil)
		return dispatcher.EndGroups
	}

	update, err := utils.SendMediaCopy(ctx, chatId, u.EffectiveMessage.Media, u.EffectiveMessage.Message.Message)

	if err != nil {
		utils.Logger.Sugar().Error(err)
		ctx.Reply(u, ext.ReplyTextString(fmt.Sprintf("Error - %s", err.Error())), nil)
		return dispatcher.EndGroups
	}
	if len(update.Updates) < 2 {
		ctx.Reply(u, ext.ReplyTextString("Error - unexpected update structure from Telegram"), nil)
		return dispatcher.EndGroups
	}
	msgIDUpdate, ok := update.Updates[0].(*tg.UpdateMessageID)
	if !ok {
		ctx.Reply(u, ext.ReplyTextString("Error - unexpected update type"), nil)
		return dispatcher.EndGroups
	}
	messageID := msgIDUpdate.ID
	newMsg, ok := update.Updates[1].(*tg.UpdateNewChannelMessage)
	if !ok {
		ctx.Reply(u, ext.ReplyTextString("Error - unexpected channel message update"), nil)
		return dispatcher.EndGroups
	}
	msg, ok := newMsg.Message.(*tg.Message)
	if !ok {
		ctx.Reply(u, ext.ReplyTextString("Error - unexpected message type"), nil)
		return dispatcher.EndGroups
	}
	doc := msg.Media
	file, err := utils.FileFromMedia(doc, msg.Message)
	if err != nil {
		ctx.Reply(u, ext.ReplyTextString(fmt.Sprintf("Error - %s", err.Error())), nil)
		return dispatcher.EndGroups
	}
	fullHash := utils.PackFile(
		file.FileName,
		file.FileSize,
		file.MimeType,
		file.ID,
	)

	metadata := utils.DetectFileMetadata(file.FileName, msg.Message)

	// Formata o nome e aplica o Fallback de segurança se ficar vazio
	displayName := utils.FormatFileNameForDisplay(metadata)
	if displayName == "" {
		displayName = file.FileName
	}

	hash := utils.GetShortHash(fullHash)
	strmFileName := utils.ProcessStrmFileName(displayName)
	strmFileNameWithExt := strmFileName + ".strm"
	linkStrm := buildStrmLink(messageID, hash, strmFileNameWithExt)
	link := fmt.Sprintf("%s/stream/%d?hash=%s", config.ValueOf.Host, messageID, hash)

	// mensagem formatada da resposta do bot, com o link para download e stream do arquivo
	messageFormatted := []styling.StyledTextOption{
		styling.Bold("🎬 Mídia Pronta para Acesso"),
		styling.Plain("\n➖➖➖➖➖➖➖➖➖➖➖\n"),
		styling.Bold("📁 Arquivo: "),
		styling.Code(file.FileName),
		styling.Plain("\n\n"),
		styling.Bold("Nome do strm: "),
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
				Text: "strm",
				URL:  linkStrm,
			},
		},
	}

	if strings.Contains(file.MimeType, "video") || strings.Contains(file.MimeType, "audio") || strings.Contains(file.MimeType, "pdf") {
		row.Buttons = append(row.Buttons, &tg.KeyboardButtonURL{
			Text: "Stream",
			URL:  link,
		})
	}
	markup := &tg.ReplyInlineMarkup{
		Rows: []tg.KeyboardButtonRow{row},
	}
	if strings.Contains(link, "http://localhost") {
		_, err = ctx.Reply(u, ext.ReplyTextStyledText(text), &ext.ReplyOpts{
			NoWebpage:        false,
			ReplyToMessageId: u.EffectiveMessage.ID,
		})
	} else {
		_, err = ctx.Reply(u, ext.ReplyTextStyledTextArray(messageFormatted), &ext.ReplyOpts{
			Markup:           markup,
			NoWebpage:        false,
			ReplyToMessageId: u.EffectiveMessage.ID,
		})
	}
	if err != nil {
		utils.Logger.Sugar().Error(err)
		ctx.Reply(u, ext.ReplyTextString(fmt.Sprintf("Error - %s", err.Error())), nil)
	}
	return dispatcher.EndGroups
}
