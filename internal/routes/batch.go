package routes

import (
	"fmt"
	"net/http"

	"EverythingSuckz/fsb/internal/utils"

	"github.com/gin-gonic/gin"
)

// LoadBatchStrm registra a rota de download do zip com os .strm de um lote.
// É descoberta automaticamente pelo mecanismo de reflection em routes.Load,
// então não precisa ser chamada manualmente em nenhum outro lugar.
func (e *allRoutes) LoadBatchStrm(r *Route) {
	log := e.log.Named("BatchStrm")
	defer log.Info("Loaded batch strm route")
	r.Engine.GET("/batch-strm/:token", getBatchStrmRoute)
}

func getBatchStrmRoute(ctx *gin.Context) {
	token := ctx.Param("token")
	if token == "" {
		http.Error(ctx.Writer, "missing token", http.StatusBadRequest)
		return
	}

	zipBytes, fileName, err := utils.GetBatchStrmZip(token)
	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusNotFound)
		return
	}

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	ctx.Data(http.StatusOK, "application/zip", zipBytes)
}
