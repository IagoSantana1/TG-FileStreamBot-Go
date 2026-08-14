package utils

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// BatchStrmEntry representa um único arquivo .strm dentro de um lote.
type BatchStrmEntry struct {
	FileName string // nome final do arquivo .strm, ex: "Filme.2024.strm"
	Content  string // conteúdo do .strm (o link de stream)
}

type batchStrmRecord struct {
	entries   []BatchStrmEntry
	name      string // nome (já sanitizado) usado para o arquivo .zip final
	createdAt time.Time
}

var (
	batchStrmMu    sync.Mutex
	batchStrmStore = map[string]*batchStrmRecord{}
)

// batchStrmTTL define por quanto tempo um zip de lote fica disponível para
// download depois de gerado (evita acúmulo indefinido em memória).
const batchStrmTTL = 24 * time.Hour

// invalidZipNameChars remove caracteres que não são seguros em nomes de
// arquivo (sistema de arquivos e cabeçalho HTTP Content-Disposition).
var invalidZipNameChars = regexp.MustCompile(`[\\/:*?"<>|]`)

// RegisterBatchStrm gera um token único, guarda as entradas em memória e
// retorna o token para montar a URL de download (/batch-strm/:token).
// seriesName é o título detectado do lote (ex: "Breaking Bad") e é usado
// para nomear o .zip; se vazio, cai para um nome genérico.
func RegisterBatchStrm(entries []BatchStrmEntry, seriesName string) (string, error) {
	if len(entries) == 0 {
		return "", fmt.Errorf("nenhum arquivo .strm para registrar")
	}

	token, err := generateBatchToken()
	if err != nil {
		return "", err
	}

	batchStrmMu.Lock()
	batchStrmStore[token] = &batchStrmRecord{
		entries:   entries,
		name:      sanitizeZipName(seriesName),
		createdAt: time.Now(),
	}
	batchStrmMu.Unlock()

	return token, nil
}

// GetBatchStrmZip monta o .zip em memória para o token informado, contendo
// um arquivo .strm por entrada registrada. Retorna os bytes do zip e o
// nome de arquivo sugerido (ex: "Breaking Bad.zip"). Retorna erro se o
// token não existir ou já tiver expirado.
func GetBatchStrmZip(token string) ([]byte, string, error) {
	batchStrmMu.Lock()
	record, ok := batchStrmStore[token]
	batchStrmMu.Unlock()

	if !ok {
		return nil, "", fmt.Errorf("token inválido ou expirado")
	}

	if time.Since(record.createdAt) > batchStrmTTL {
		batchStrmMu.Lock()
		delete(batchStrmStore, token)
		batchStrmMu.Unlock()
		return nil, "", fmt.Errorf("link expirado")
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	usedNames := make(map[string]int)

	for _, entry := range record.entries {
		name := uniqueZipName(usedNames, entry.FileName)

		w, err := zw.Create(name)
		if err != nil {
			zw.Close()
			return nil, "", err
		}
		if _, err := w.Write([]byte(entry.Content)); err != nil {
			zw.Close()
			return nil, "", err
		}
	}

	if err := zw.Close(); err != nil {
		return nil, "", err
	}

	fileName := record.name
	if fileName == "" {
		fileName = "arquivos_" + token[:8]
	}

	return buf.Bytes(), fileName + ".zip", nil
}

// CleanupExpiredBatchStrm remove tokens vencidos da memória. Pode ser
// chamada periodicamente (ex: goroutine com time.Ticker no bootstrap do
// bot) para evitar crescimento indefinido do mapa em instâncias de longa
// duração.
func CleanupExpiredBatchStrm() {
	batchStrmMu.Lock()
	defer batchStrmMu.Unlock()
	for token, record := range batchStrmStore {
		if time.Since(record.createdAt) > batchStrmTTL {
			delete(batchStrmStore, token)
		}
	}
}

// sanitizeZipName limpa o título detectado (ex: "Breaking Bad") para que
// possa ser usado com segurança como nome de arquivo .zip.
func sanitizeZipName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	name = invalidZipNameChars.ReplaceAllString(name, "")
	name = strings.Join(strings.Fields(name), " ") // colapsa espaços múltiplos

	const maxLen = 80
	if len(name) > maxLen {
		name = strings.TrimSpace(name[:maxLen])
	}

	return name
}

// uniqueZipName evita colisão de nomes dentro do zip quando dois arquivos
// do mesmo lote tem o mesmo nome final (ex: "Episódio.strm" duplicado).
func uniqueZipName(used map[string]int, name string) string {
	if name == "" {
		name = "arquivo.strm"
	}

	count := used[name]
	used[name] = count + 1

	if count == 0 {
		return name
	}

	ext := ""
	base := name
	if idx := lastDotIndex(name); idx != -1 {
		ext = name[idx:]
		base = name[:idx]
	}

	return fmt.Sprintf("%s (%d)%s", base, count, ext)
}

func lastDotIndex(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i
		}
	}
	return -1
}

func generateBatchToken() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
