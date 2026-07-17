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

	// Passo 1: Combinar informações do nome + descrição
	metadata.ExpandedName = combineNameAndDescription(fileName, messageDescription)

	// Passo 2: Extrair informações básicas (título, qualidade)
	metadata.Title = extractTitle(fileName, messageDescription)
	metadata.Quality = ExtractQualityFromDescription(messageDescription)

	// Passo 3: Detectar tipo (filme/série)
	metadata.Type, metadata.TypeConfidence = detectFileType(metadata.ExpandedName)

	// Passo 4: Extrair temporada e episódio
	if metadata.Type == "series" {
		season, episode, hasSeasonEpisode := extractSeasonEpisode(metadata.ExpandedName)
		metadata.Season = season
		metadata.Episode = episode

		// Determina o que está faltando
		if !hasSeasonEpisode {
			if metadata.Episode == 0 {
				metadata.MissingInfo = append(metadata.MissingInfo, "episode")
			}
			if metadata.Season == 0 {
				metadata.MissingInfo = append(metadata.MissingInfo, "season")
			}
		}
	}

	// Passo 5: Verificar se tem tudo
	metadata.AllInfoFound = len(metadata.MissingInfo) == 0

	return metadata
}

func combineNameAndDescription(fileName, description string) string {
	cleanName := strings.TrimSpace(fileName)
	cleanName = removeExtensions(cleanName)
	cleanName = removeAtMentions(cleanName)
	cleanName = strings.TrimSpace(cleanName)

	cleanDesc := strings.TrimSpace(description)
	if cleanDesc == "" {
		return cleanName
	}

	cleanDesc = reSiga.ReplaceAllString(cleanDesc, "")
	cleanDesc = reAtMention.ReplaceAllString(cleanDesc, "")
	cleanDesc = removeAtMentions(cleanDesc)
	cleanDesc = strings.TrimSpace(cleanDesc)

	// Se a própria legenda contiver dados de ep, ela vira prioridade
	reCheck := regexp.MustCompile(`(?i)([ST]\d+|[0-9]+x[0-9]+)`)
	if reCheck.MatchString(cleanDesc) {
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
	// Padrões de ep locais e robustos para classificação de tipo
	reSE := regexp.MustCompile(`(?i)[ST](\d+)[EX\s._\-]*(\d+)`)
	reX := regexp.MustCompile(`(?i)\b(\d+)[xX](\d+)\b`)

	if reSE.MatchString(expandedName) || reX.MatchString(expandedName) {
		return "series", 95
	}

	reEpIsolated := regexp.MustCompile(`(?i)\b(EP|E|EPISODIO)[._\-\s]*\d+\b`)
	if reEpIsolated.MatchString(expandedName) {
		return "series", 80
	}

	if regexp.MustCompile(`\b(19|20)\d{2}\b`).MatchString(expandedName) {
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

// extractSeasonEpisode extrai cirurgicamente APENAS os dígitos numéricos
func extractSeasonEpisode(expandedName string) (season int, episode int, hasInfo bool) {
	// Grupos de captura (\d+) pegam apenas os números isolados!
	reSE := regexp.MustCompile(`(?i)[ST](\d+)[EX\s._\-]*(\d+)`)
	reX := regexp.MustCompile(`(?i)\b(\d+)[xX](\d+)\b`)

	if matches := reSE.FindStringSubmatch(expandedName); len(matches) >= 3 {
		season, _ = strconv.Atoi(matches[1])  // Captura apenas "04" -> vira 4
		episode, _ = strconv.Atoi(matches[2]) // Captura apenas "02" -> vira 2
		hasInfo = true
		return
	}

	if matches := reX.FindStringSubmatch(expandedName); len(matches) >= 3 {
		season, _ = strconv.Atoi(matches[1])
		episode, _ = strconv.Atoi(matches[2])
		hasInfo = true
		return
	}

	if seasonNum, found := ExtractSeasonNumber(expandedName); found && seasonNum > 0 {
		season = seasonNum
	}

	reEpIsolated := regexp.MustCompile(`(?i)\b(?:EP|E|EPISODIO)[._\-\s]*(\d+)\b`)
	if matches := reEpIsolated.FindStringSubmatch(expandedName); len(matches) >= 2 {
		episode, _ = strconv.Atoi(matches[1])
		hasInfo = true
	}

	if season > 0 || episode > 0 {
		hasInfo = true
	}

	return
}

func extractTitle(fileName, description string) string {
	cleanName := strings.TrimSpace(fileName)
	cleanName = removeExtensions(cleanName)
	cleanName = removeAtMentions(cleanName)
	cleanName = strings.TrimSpace(cleanName)

	if cleanName != "" && !IsGenericFileName(cleanName) {
		// 1. Remove os padrões de episódio (ex: S02E06)
		title := removeEpisodePatterns(cleanName)

		// 2. Remove resoluções e fontes residuais para o título não ficar poluído
		reCleanExtra := regexp.MustCompile(`(?i)\b(2160p|1080p|720p|480p|4k|2k|WEB-?DL|BluRay|NF-?WEB-?DL|AMZN-?WEB-?DL|BRRip|WEBRip|HDRip|DVDRip)\b`)
		title = reCleanExtra.ReplaceAllString(title, "")
		title = removeAtMentions(title)

		// 3. CRÍTICO: Substitui múltiplos sublinhados, pontos ou traços por espaços limpos
		// (Para o DisplayName do Telegram, espaços tornam a leitura muito mais elegante)
		title = regexp.MustCompile(`[_\.\-\s]+`).ReplaceAllString(title, " ")
		title = strings.TrimSpace(title)

		if title != "" {
			return title
		}
	}

	if description != "" {
		cleanDesc := strings.TrimSpace(description)
		cleanDesc = reSiga.ReplaceAllString(cleanDesc, "")
		cleanDesc = reAtMention.ReplaceAllString(cleanDesc, "")
		cleanDesc = removeAtMentions(cleanDesc)
		cleanDesc = strings.TrimSpace(cleanDesc)

		if idx := strings.IndexAny(cleanDesc, "\n|.!?"); idx > 0 {
			cleanDesc = cleanDesc[:idx]
		}
		cleanDesc = strings.TrimSpace(cleanDesc)

		title := removeEpisodePatterns(cleanDesc)
		title = removeYearPatterns(title)
		title = removeAtMentions(title)

		// Normaliza os separadores da descrição também
		title = regexp.MustCompile(`[_\.\-\s]+`).ReplaceAllString(title, " ")
		title = strings.TrimSpace(title)
		return title
	}

	return ""
}

func removeAtMentions(text string) string {
	reAtMentionsClean := regexp.MustCompile(`@[\w_]+`)
	return reAtMentionsClean.ReplaceAllString(text, "")
}

func removeEpisodePatterns(text string) string {
	reSE := regexp.MustCompile(`(?i)[ST]\d+[EX\s._\-]*\d+`)
	reX := regexp.MustCompile(`(?i)\b\d+[xX]\d+\b`)
	reEp := regexp.MustCompile(`(?i)\b(EP|E|EPISODIO)[._\-\s]*\d+\b`)

	text = reSE.ReplaceAllString(text, "")
	text = reX.ReplaceAllString(text, "")
	text = reEp.ReplaceAllString(text, "")
	text = regexp.MustCompile(`(?i)\bseason\s*\d+\b`).ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}

func removeYearPatterns(text string) string {
	return strings.TrimSpace(regexp.MustCompile(`[\(\[\s_]+(?:19|20)\d{2}[\)\]\s_]*`).ReplaceAllString(text, ""))
}

func matchesPattern(text, pattern string) bool {
	re := regexp.MustCompile(pattern)
	return re.MatchString(text)
}

func DetermineQuestionsNeeded(metadata *FileMetadata) []string {
	questions := []string{}
	if metadata.Type == "unknown" {
		questions = append(questions, "type")
	}
	if IsGenericFileName(metadata.OriginalFileName) && metadata.Title == "" {
		questions = append(questions, "title")
	}
	if metadata.Type == "series" && metadata.Season == 0 {
		questions = append(questions, "season")
	}
	if metadata.Type == "series" && metadata.Episode > 0 && metadata.Season == 0 {
		if !Contains(questions, "season") {
			questions = append(questions, "season")
		}
	}
	return questions
}

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
