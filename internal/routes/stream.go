package routes

import (
	"EverythingSuckz/fsb/internal/bot"
	"EverythingSuckz/fsb/internal/stream"
	"EverythingSuckz/fsb/internal/types"
	"EverythingSuckz/fsb/internal/utils"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gotd/td/tg"
	range_parser "github.com/quantumsheep/range-parser"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

var log *zap.Logger

func (e *allRoutes) LoadHome(r *Route) {
	log = e.log.Named("Stream")
	defer log.Info("Loaded stream route")
	r.Engine.GET("/stream/:messageID", getStreamRoute)
	r.Engine.GET("/strm/:messageID", getStrmRoute)
}

func getStreamRoute(ctx *gin.Context) {
	w := ctx.Writer
	r := ctx.Request

	messageIDParm := ctx.Param("messageID")
	messageID, err := strconv.Atoi(messageIDParm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authHash := ctx.Query("hash")
	if authHash == "" {
		http.Error(w, "missing hash param", http.StatusBadRequest)
		return
	}

	worker := bot.GetNextWorker()

	file, err := utils.TimeFuncWithResult(log, "FileFromMessage", func() (*types.File, error) {
		return utils.FileFromMessage(ctx, worker.Client, messageID)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	expectedHash := utils.PackFile(
		file.FileName,
		file.FileSize,
		file.MimeType,
		file.ID,
	)
	if !utils.CheckHash(authHash, expectedHash) {
		http.Error(w, "invalid hash", http.StatusBadRequest)
		return
	}

	// for photo messages
	if file.FileSize == 0 {
		res, err := worker.Client.API().UploadGetFile(ctx, &tg.UploadGetFileRequest{
			Location: file.Location,
			Offset:   0,
			Limit:    1024 * 1024,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result, ok := res.(*tg.UploadFile)
		if !ok {
			http.Error(w, "unexpected response", http.StatusInternalServerError)
			return
		}
		fileBytes := result.GetBytes()
		ctx.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", file.FileName))
		if r.Method != "HEAD" {
			ctx.Data(http.StatusOK, file.MimeType, fileBytes)
		}
		return
	}

	ctx.Header("Accept-Ranges", "bytes")
	var start, end int64
	rangeHeader := r.Header.Get("Range")

	if rangeHeader == "" {
		start = 0
		end = file.FileSize - 1
		w.WriteHeader(http.StatusOK)
	} else {
		ranges, err := range_parser.Parse(file.FileSize, r.Header.Get("Range"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		start = ranges[0].Start
		end = ranges[0].End
		ctx.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, file.FileSize))
		log.Info("Content-Range", zap.Int64("start", start), zap.Int64("end", end), zap.Int64("fileSize", file.FileSize))
		w.WriteHeader(http.StatusPartialContent)
	}

	contentLength := end - start + 1
	mimeType := file.MimeType

	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	ctx.Header("Content-Type", mimeType)
	ctx.Header("Content-Length", strconv.FormatInt(contentLength, 10))

	disposition := "inline"

	if ctx.Query("d") == "true" {
		disposition = "attachment"
	}

	ctx.Header("Content-Disposition", fmt.Sprintf("%s; filename=\"%s\"", disposition, file.FileName))

	if r.Method != "HEAD" {
		pipe, err := stream.NewStreamPipe(ctx, worker.Client, file.Location, start, end, log)
		if err != nil {
			log.Error("Failed to create stream pipe", zap.Error(err))
			return
		}
		defer pipe.Close()
		if _, err := io.CopyN(w, pipe, contentLength); err != nil {
			if !utils.IsClientDisconnectError(err) {
				log.Error("Error while copying stream", zap.Error(err))
			}
		}
	}
}

func getStrmRoute(ctx *gin.Context) {
	writer := ctx.Writer
	request := ctx.Request

	messageIDParams := ctx.Param("messageID")
	messageID, err := strconv.Atoi(messageIDParams)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	authHash := ctx.Query("hash")
	if authHash == "" {
		http.Error(writer, "missing hash param", http.StatusBadRequest)
		return
	}

	worker := bot.GetNextWorker()

	file, err := utils.FileFromMessage(ctx, worker.Client, messageID)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = os.Getenv("APP_URL")
	}
	if host == "" {
		host = fmt.Sprintf("http://localhost:%s", os.Getenv("PORT"))
	}

	// 2. Garante que o host tenha o protocolo (http:// ou https://)
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		proto := "http"
		// Se o Codespace ou Proxy encaminhar como HTTPS, respeita isso
		if request.TLS != nil || request.Header.Get("X-Forwarded-Proto") == "https" {
			proto = "https"
		}
		host = fmt.Sprintf("%s://%s", proto, host)
	}

	// Remove barra final se o usuário tiver configurado por engano (evita dupla barra no link)
	host = strings.TrimSuffix(host, "/")

	// Gera o link de download
	downloadLink := fmt.Sprintf("%s/stream/%d?hash=%s", host, messageID, authHash)

	// Processa o nome do arquivo .strm usando o DisplayName melhorado
	nameParam := ctx.Query("name")
	var strmFileName string
	if nameParam != "" {
		decoded, err := url.QueryUnescape(nameParam)
		if err == nil && strings.TrimSpace(decoded) != "" {
			strmFileName = decoded
		}
	}

	if strmFileName == "" {
		// Usar DisplayName se disponível, senão usar FileName original
		nameToProcess := file.DisplayName
		if nameToProcess == "" {
			nameToProcess = file.FileName
		}
		processedFileName := utils.ProcessStrmFileName(nameToProcess)
		strmFileName = processedFileName + ".strm"
	}

	// Define headers para download do arquivo .strm
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", strmFileName))
	ctx.Header("Content-Length", strconv.Itoa(len(downloadLink)))

	// Envia o conteúdo do arquivo .strm (apenas o link de download)
	ctx.String(http.StatusOK, downloadLink)
}
