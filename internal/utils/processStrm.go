package utils

import (
	"regexp"
	"strings"
	"unicode"
)

func ProcessStrmFileName(fileName string) string {
	return ProcessStrmFileNameWithQuality(fileName, 0)
}

func ProcessStrmFileNameWithQuality(fileName string, fileSize int64) string {
	if fileName == "" {
		return fileName
	}

	// 0. Remove símbolos Unicode (como ●) e emojis
	fileName = cleanNameForDownload(fileName)

	// Remove domínios (como .eu, .com) e nomes de sites de pirataria
	fileName = reDomains.ReplaceAllString(fileName, "")
	fileName = reSiteNames.ReplaceAllString(fileName, "")

	// 1. Troca sublinhados e quebras de linha por espaços e limpa extensões
	fileName = strings.ReplaceAll(fileName, "_", " ")
	fileName = strings.ReplaceAll(fileName, "\n", " ")
	fileName = strings.TrimSpace(fileName)
	fileName = reSiga.ReplaceAllString(fileName, "")
	fileName = reAtMention.ReplaceAllString(fileName, "")
	fileName = reExtension.ReplaceAllString(fileName, "")

	// Extrai a qualidade logo no início para usar como "Versão" no padrão Jellyfin
	qualityInfo := extractQualityInfo(fileName)

	// 2. Normaliza qualquer padrão de temporada/episódio para SXXEXX
	fileName = NormalizeEpisodeFormat(fileName)

	// 3. SE FOR SÉRIE (contém SXXEXX)
	if matches := reSeasonEpisodePattern.FindStringIndex(fileName); len(matches) == 2 {
		rawTitle := fileName[:matches[0]]

		// Pega o bloco exato onde a temporada e episódio foram encontrados
		rawSE := fileName[matches[0]:matches[1]]

		// Força o formato "S01E01"
		seasonEpisode := regexp.MustCompile(`[._\s]`).ReplaceAllString(strings.ToUpper(rawSE), "")

		// Limpa lixo do título
		cleanTitle := reJunkTags.ReplaceAllString(rawTitle, "")
		cleanTitle = reYearWrapped.ReplaceAllString(cleanTitle, "")
		cleanTitle = reYearPlain.ReplaceAllString(cleanTitle, "")
		cleanTitle = reParensBrackets.ReplaceAllString(cleanTitle, "")
		cleanTitle = regexp.MustCompile(`\b\d+\b`).ReplaceAllString(cleanTitle, "")

		// Substitui múltiplos espaços/pontos/traços por um único espaço
		cleanTitle = regexp.MustCompile(`[_\.\-\s]+`).ReplaceAllString(cleanTitle, " ")
		cleanTitle = strings.TrimSpace(cleanTitle)

		// PADRÃO JELLYFIN SÉRIES: "Nome da Serie S01E01 - 1080p"
		if qualityInfo != "" {
			return cleanTitle + " " + seasonEpisode + " - " + qualityInfo
		}
		return cleanTitle + " " + seasonEpisode
	}

	// 4. SE FOR FILME: Mantém o fluxo normal
	fileName = reYearWrapped.ReplaceAllString(fileName, "")
	fileName = reYearPlain.ReplaceAllString(fileName, "")
	fileName = reResolution.ReplaceAllString(fileName, "")
	fileName = reParensBrackets.ReplaceAllString(fileName, "")
	fileName = reQualityCodecs.ReplaceAllString(fileName, "")
	fileName = reReleaseGroup.ReplaceAllString(fileName, "")
	fileName = reJunkTags.ReplaceAllString(fileName, "")

	// Substitui múltiplos espaços/pontos/traços por um único espaço
	fileName = regexp.MustCompile(`[_\.\-\s]+`).ReplaceAllString(fileName, " ")
	fileName = strings.TrimSpace(fileName)

	// PADRÃO JELLYFIN FILMES: "Nome do Filme - 1080p"
	if qualityInfo != "" {
		fileName = fileName + " - " + qualityInfo
	}

	return fileName
}

// extractQualityInfo extrai APENAS a resolução (1080p, 720p, 4K, etc) do nome do arquivo
func extractQualityInfo(fileName string) string {
	// Procura por resoluções padrão do Jellyfin
	reRes := regexp.MustCompile(`(?i)\b(2160p|1080p|720p|480p|360p|4k)\b`)

	if matches := reRes.FindStringSubmatch(fileName); len(matches) > 1 {
		res := strings.ToLower(matches[1])
		// Se for 4k, deixa o K maiúsculo para ficar mais bonito no Jellyfin
		if res == "4k" {
			return "4K"
		}
		return res
	}

	return ""
}

// removeDuplicateEpisodes remove episódios duplicados mantendo apenas o último e a resolução
func removeDuplicateEpisodes(fileName string) string {
	matches := reSeasonEpisodePattern.FindAllStringIndex(fileName, -1)
	if len(matches) <= 1 {
		return fileName
	}

	firstMatch := matches[0]
	lastMatch := matches[len(matches)-1]

	// Pega o nome base (antes do primeiro episódio duplicado)
	beforeFirstEpisode := fileName[:firstMatch[0]]
	beforeFirstEpisode = strings.TrimRight(beforeFirstEpisode, "._- ")

	// Pega o último episódio (o correto)
	lastEpisode := fileName[lastMatch[0]:lastMatch[1]]

	// Puxa a resolução do arquivo original
	quality := extractQualityInfo(fileName)

	// Remonta de forma limpa
	result := strings.TrimSpace(beforeFirstEpisode)

	if len(result) > 0 {
		result += " "
	}
	result += lastEpisode

	if quality != "" {
		result += " - " + quality
	}

	return result
}

// extractQualityInfoFromString faz a mesma coisa para strings específicas
func extractQualityInfoFromString(s string) string {
	return extractQualityInfo(s)
}

// cleanNameForDownload remove emojis e caracteres problemáticos
func cleanNameForDownload(text string) string {
	if text == "" {
		return text
	}

	text = strings.Map(func(r rune) rune {
		if unicode.IsSymbol(r) || r > 0x1F000 || unicode.IsMark(r) {
			return -1
		}
		return r
	}, text)

	problematicChars := []string{":", "<", ">", "?", "\"", "|"}
	for _, char := range problematicChars {
		text = strings.ReplaceAll(text, char, "")
	}

	return text
}

func IsGenericFileName(fileName string) bool {
	if fileName == "" {
		return true
	}

	nameWithoutExt := fileName
	re := regexp.MustCompile(`\.[a-zA-Z0-9]+$`)
	nameWithoutExt = re.ReplaceAllString(fileName, "")

	lowerName := strings.ToLower(strings.TrimSpace(nameWithoutExt))

	genericNames := map[string]bool{
		"file": true, "arquivo": true, "document": true,
		"documento": true, "video": true, "vídeo": true,
		"audio": true, "áudio": true, "image": true,
		"imagem": true, "photo": true, "foto": true,
		"media": true, "mídia": true, "download": true,
		"unnamed": true, "untitled": true, "noname": true,
		"unknown": true, "desconhecido": true, "sem nome": true,
	}

	if genericNames[lowerName] {
		return true
	}

	baseName := lowerName
	baseName = regexp.MustCompile(`(?i)\s*-?\s*s\d+e\d+.*$`).ReplaceAllString(baseName, "")
	baseName = regexp.MustCompile(`(?i)\s*-?\s*\d+x\d+.*$`).ReplaceAllString(baseName, "")
	baseName = strings.TrimSpace(baseName)

	sitePatterns := []string{
		"baixar.*mp4", "download.*mp4", "filmesmp4", "seriesmp4",
		"comandotorrents", "comandofilmes", "torrentsbr", "bludv",
		"mega.*filmes", "series.*online", "CajaréTorrents", "cajarétorrents",
	}

	for _, pattern := range sitePatterns {
		if matched, _ := regexp.MatchString("(?i)"+pattern, baseName); matched {
			cleaned := regexp.MustCompile("(?i)"+pattern).ReplaceAllString(baseName, "")
			cleaned = regexp.MustCompile(`[_\.\-\s]+`).ReplaceAllString(cleaned, "")
			if len(cleaned) < 3 {
				return true
			}
		}
	}

	return false
}

// ExtractQualityFromDescription extrai APENAS a resolução da descrição da mensagem
func ExtractQualityFromDescription(description string) string {
	if description == "" {
		return ""
	}

	qualityPatterns := []struct {
		pattern string
		label   string
	}{
		{`(?i)\b(2160p|4K)\b`, "4K"},
		{`(?i)\b(1080p|1080|Full.?HD|FHD)\b`, "1080p"},
		{`(?i)\b(720p|720|HD)\b`, "720p"},
		{`(?i)\b(480p|SD)\b`, "480p"},
	}

	for _, q := range qualityPatterns {
		re := regexp.MustCompile(q.pattern)
		if re.MatchString(description) {
			return q.label
		}
	}

	return ""
}
