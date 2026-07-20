package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// FileMetadata representa todos os metadados extraídos de um arquivo e sua descrição
type FileMetadata struct {
	OriginalFileName string   // Nome original do arquivo
	MessageText      string   // Texto da mensagem/descrição
	Type             string   // "movie", "series", "unknown"
	TypeConfidence   int      // 0-100: confiança na detecção
	Title            string   // Título da obra
	Season           int      // 0 = não encontrado
	Episode          int      // 0 = não encontrado
	Quality          string   // Qualidade extraída
	AllInfoFound     bool     // Se encontrou tudo necessário
	MissingInfo      []string // Informações faltantes: "title", "season", "episode"
	ExpandedName     string   // Nome combinado (arquivo + descrição processada)
}

// DetectFileMetadata analisa arquivo + descrição para extrair metadados inteligentemente
func DetectFileMetadata(fileName, messageDescription string) *FileMetadata {
	metadata := &FileMetadata{
		OriginalFileName: fileName,
		MessageText:      messageDescription,
		Season:           0,
		Episode:          0,
		TypeConfidence:   0,
		MissingInfo:      []string{},
	}

	metadata.ExpandedName = combineNameAndDescription(fileName, messageDescription)
	metadata.Title = extractTitle(fileName, messageDescription)
	metadata.Quality = ExtractQualityFromDescription(messageDescription)
	metadata.Type, metadata.TypeConfidence = detectFileType(metadata.ExpandedName)

	if metadata.Type == "series" {
		season, episode, hasSeasonEpisode := extractSeasonEpisode(metadata.ExpandedName)
		metadata.Season = season
		metadata.Episode = episode

		if !hasSeasonEpisode {
			if metadata.Episode == 0 {
				metadata.MissingInfo = append(metadata.MissingInfo, "episode")
			}
			if metadata.Season == 0 {
				metadata.MissingInfo = append(metadata.MissingInfo, "season")
			}
		}
	}

	metadata.AllInfoFound = len(metadata.MissingInfo) == 0
	return metadata
}

func combineNameAndDescription(fileName, description string) string {
	cleanName := strings.ReplaceAll(fileName, "_", " ")
	cleanName = strings.TrimSpace(cleanName)
	cleanName = removeExtensions(cleanName)
	cleanName = removeAtMentions(cleanName)

	cleanDesc := strings.ReplaceAll(description, "_", " ")
	cleanDesc = strings.TrimSpace(cleanDesc)
	if cleanDesc == "" {
		return cleanName
	}

	cleanDesc = reSiga.ReplaceAllString(cleanDesc, "")
	cleanDesc = reAtMention.ReplaceAllString(cleanDesc, "")
	cleanDesc = strings.TrimSpace(cleanDesc)

	if reSeasonPatternUnified.MatchString(cleanDesc) || rePtSeasonPattern.MatchString(cleanDesc) || reUnifiedEpisodeOnly.MatchString(cleanDesc) {
		if cleanName != "" && !isGenericOrEmpty(cleanName) {
			return cleanDesc + " " + cleanName
		}
		return cleanDesc
	}

	if isGenericOrEmpty(cleanName) {
		return cleanDesc
	}

	return cleanDesc + " " + cleanName
}

func removeExtensions(fileName string) string {
	return reExtension.ReplaceAllString(fileName, "")
}

func isGenericOrEmpty(name string) bool {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return true
	}
	return IsGenericFileName(trimmed)
}

func detectFileType(expandedName string) (string, int) {
	if reSeasonPatternUnified.MatchString(expandedName) || rePtSeasonPattern.MatchString(expandedName) {
		return "series", 95
	}

	if reUnifiedEpisodeOnly.MatchString(expandedName) {
		return "series", 80
	}

	if reYearPattern.MatchString(expandedName) {
		return "movie", 85
	}

	if matchesPattern(expandedName, `(?i)\bfilme\b|\bmovie\b`) {
		return "movie", 75
	}

	if matchesPattern(expandedName, `(?i)\bserie|série\b|\btemporada\b`) {
		return "series", 75
	}

	return "unknown", 0
}

// 🚀 FIX: Agora extrai corretamente Temporada e Episódio de formatos em português
func extractSeasonEpisode(expandedName string) (season int, episode int, hasInfo bool) {
	clean := strings.ReplaceAll(expandedName, "_", " ")

	// 1. Tenta extrair de "Temporada 01 episódio 08"
	if matches := rePtSeasonEpisodeCapture.FindStringSubmatch(clean); len(matches) >= 3 {
		season, _ = strconv.Atoi(matches[1])
		episode, _ = strconv.Atoi(matches[2])
		hasInfo = true
		return
	}

	// 2. Tenta extrair de S01E08 ou 1x08
	if matches := reUnifiedEpisode.FindStringSubmatch(clean); len(matches) == 5 {
		if matches[1] != "" {
			season, _ = strconv.Atoi(matches[1])
			episode, _ = strconv.Atoi(matches[2])
			hasInfo = true
			return
		}
		if matches[3] != "" {
			season, _ = strconv.Atoi(matches[3])
			episode, _ = strconv.Atoi(matches[4])
			hasInfo = true
			return
		}
	}

	if seasonNum, found := ExtractSeasonNumber(clean); found && seasonNum > 0 {
		season = seasonNum
	}

	if matches := reUnifiedEpisodeOnly.FindStringSubmatch(clean); len(matches) == 3 {
		episode, _ = strconv.Atoi(matches[2])
		hasInfo = true
	}

	if season > 0 || episode > 0 {
		hasInfo = true
	}

	return
}

func extractTitle(fileName, description string) string {
	cleanName := strings.ReplaceAll(fileName, "_", " ")
	cleanName = strings.TrimSpace(cleanName)
	cleanName = removeExtensions(cleanName)
	cleanName = removeAtMentions(cleanName)

	if cleanName != "" && !IsGenericFileName(cleanName) {
		title := cleanName

		// 🚀 CORREÇÃO CRÍTICA: Encontra a primeira ocorrência do padrão de episódio
		// e CORRTA TUDO que vem do episódio para a frente!
		firstMatchIdx := -1
		for _, re := range []*regexp.Regexp{
			rePtSeasonEpisodeCapture,
			reUnifiedEpisode,
			reUnifiedEpisodeOnly,
			reSeasonNum,
		} {
			if loc := re.FindStringIndex(title); len(loc) == 2 {
				if firstMatchIdx == -1 || loc[0] < firstMatchIdx {
					firstMatchIdx = loc[0]
				}
			}
		}

		// Se encontrou o episódio/temporada, pega estritamente o texto ANTES dele
		if firstMatchIdx > 0 {
			title = title[:firstMatchIdx]
		}

		// Limpa eventuais tags residuais que estivessem ANTES do episódio (ex: resoluções)
		title = reJunkTags.ReplaceAllString(title, " ")
		title = removeAtMentions(title)

		// Normaliza múltiplos espaços e separadores
		title = regexp.MustCompile(`[_\.\-\s]+`).ReplaceAllString(title, " ")
		title = strings.TrimSpace(title)

		if title != "" {
			return title
		}
	}

	// Caso o nome do arquivo seja genérico, faz a mesma limpeza na legenda/descrição
	if description != "" {
		cleanDesc := strings.ReplaceAll(description, "_", " ")
		cleanDesc = strings.TrimSpace(cleanDesc)
		cleanDesc = reSiga.ReplaceAllString(cleanDesc, "")
		cleanDesc = reAtMention.ReplaceAllString(cleanDesc, "")

		if idx := strings.IndexAny(cleanDesc, "\n|.!?"); idx > 0 {
			cleanDesc = cleanDesc[:idx]
		}
		cleanDesc = strings.TrimSpace(cleanDesc)

		firstMatchIdx := -1
		for _, re := range []*regexp.Regexp{
			rePtSeasonEpisodeCapture,
			reUnifiedEpisode,
			reUnifiedEpisodeOnly,
			reSeasonNum,
		} {
			if loc := re.FindStringIndex(cleanDesc); len(loc) == 2 {
				if firstMatchIdx == -1 || loc[0] < firstMatchIdx {
					firstMatchIdx = loc[0]
				}
			}
		}

		if firstMatchIdx > 0 {
			cleanDesc = cleanDesc[:firstMatchIdx]
		}

		title := reJunkTags.ReplaceAllString(cleanDesc, " ")
		title = removeYearPatterns(title)
		title = removeAtMentions(title)

		title = regexp.MustCompile(`[_\.\-\s]+`).ReplaceAllString(title, " ")
		title = strings.TrimSpace(title)
		return title
	}

	return ""
}

func removeAtMentions(text string) string {
	return reAtMention.ReplaceAllString(text, "")
}

// 🚀 FIX: Substitui padrões de episódio por um ESPAÇO (" ") para nunca colar palavras
func removeEpisodePatterns(text string) string {
	text = rePtSeasonEpisodeCapture.ReplaceAllString(text, " ")
	text = reUnifiedEpisode.ReplaceAllString(text, " ")
	text = reUnifiedEpisodeOnly.ReplaceAllString(text, " ")
	text = reSeasonNum.ReplaceAllString(text, " ")
	text = reShortSeason.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func removeYearPatterns(text string) string {
	return strings.TrimSpace(regexp.MustCompile(`[\(\[\s_]+(19|20)\d{2}[\)\]\s_]*`).ReplaceAllString(text, " "))
}

func matchesPattern(text, pattern string) bool {
	re := regexp.MustCompile(pattern)
	return re.MatchString(text)
}

// FormatFileNameForDisplay formata o nome do arquivo para exibição limpa no Telegram
func FormatFileNameForDisplay(metadata *FileMetadata) string {
	title := removeAtMentions(metadata.Title)
	title = strings.TrimSpace(title)

	if metadata.Type == "series" && metadata.Season > 0 && metadata.Episode > 0 {
		return fmt.Sprintf("%s - S%02dE%02d", title, metadata.Season, metadata.Episode)
	}

	if metadata.Type == "series" && metadata.Season > 0 {
		return fmt.Sprintf("%s - S%02d", title, metadata.Season)
	}

	if metadata.Type == "movie" {
		if metadata.Quality != "" {
			return fmt.Sprintf("%s.%s", title, metadata.Quality)
		}
		return title
	}

	return title
}
