package utils

import (
	"regexp"
	"strings"
)

var (
	reExtensions      = regexp.MustCompile(`(?i)\.(mp4|mkv|avi|mp3|flac|pdf|strm)$`) // Sua regex atual de extensões
	reAtMentionsClean = regexp.MustCompile(`@[\w_]+`)
	reMovieKeyword    = regexp.MustCompile(`(?i)\b(filme|movie)\b`)
	reSeriesKeyword   = regexp.MustCompile(`(?i)\b(serie|série|temporada)\b`)
)

type fileMetadata struct {
	OriginalFileName string
	MessageText      string
	Type             string
	TypeConfidence   int
	Title            string
	Season           int
	Episode          int
	Quality          string
	AllInfoFound     bool
	MissingInfo      []string
	ExpandedName     string
}

// DetectFileMetadata analisa arquivo + descrição para extrair metadados inteligentemente
// Retorna uma estrutura com todas as informações detectadas e quais estão faltando
func DetectFileMetadata(fileName, messageDescription string) *fileMetadata {
	metadata := &fileMetadata{
		OriginalFileName: fileName,
		MessageText:      messageDescription,
		Season:           0,
		Episode:          0,
		TypeConfidence:   0,
		MissingInfo:      []string{},
	}

	return metadata
}

func extractFileName(fileName, description string) string {
	cleanName := strings.TrimSpace(fileName)
	cleanName = cleanFileName(cleanName)
	cleanName = strings.TrimSpace(cleanName)

	// Se nome do arquivo é específico (não genérico), usar ele

	// Se nome do arquivo é genérico, tentar extrair do texto da mensagem

	return cleanName
}

func cleanFileName(fileName string) string {
	withoutMentions := reAtMentionsClean.ReplaceAllString(fileName, "")

	return reExtensions.ReplaceAllString(withoutMentions, "")
}

func isGenericOrEmpty(name string) bool {
	trimmed := strings.TrimSpace(name)
	return trimmed == ""
}

func removeAtMentions(text string) string {
	// Remove @usuario e tudo depois até espaço ou fim
	reAtMentionsClean := regexp.MustCompile(`@[\w_]+`)
	return reAtMentionsClean.ReplaceAllString(text, "")
}

func FormatFileNameForDisplay(metadata string) string {
	title := removeAtMentions(metadata)
	title = strings.TrimSpace(title)

	// Padrão: retornar o título sem @mentions
	return title
}
