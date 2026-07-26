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

	"github.com/gotd/td/tg"
)

type pendingConfirm struct{}

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

	// Responde apenas em conversas privadas
	if peerChatId.Type != int(storage.TypeUser) {
		return dispatcher.EndGroups
	}

	// Verifica se o usuário está autorizado
	if len(config.ValueOf.AllowedUsers) != 0 &&
		!utils.Contains(config.ValueOf.AllowedUsers, chatId) {

		ctx.Reply(u,
			ext.ReplyTextString("You are not allowed to use this bot."),
			nil,
		)
		return dispatcher.EndGroups
	}

	// Verifica se é uma mídia suportada
	supported, err := supportedMediaFilter(u.EffectiveMessage)
	if err != nil {
		return err
	}

	if !supported {
		ctx.Reply(u,
			ext.ReplyTextString("Desculpe, este tipo de mensagem não é suportado."),
			nil,
		)
		return dispatcher.EndGroups
	}

	// Adiciona a mídia ao lote
	enqueueBatch(ctx, u, chatId)

	// Não processa agora.
	// O processamento acontecerá quando o lote for fechado.
	return dispatcher.EndGroups
}
