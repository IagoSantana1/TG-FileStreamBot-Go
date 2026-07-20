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

	// 1. Troca sublinhados por espaços e limpa quebras/extensões
	fileName = strings.ReplaceAll(fileName, "_", " ")
	lines := strings.Split(fileName, "\n")
	fileName = strings.TrimSpace(lines[0])
	fileName = reSiga.ReplaceAllString(fileName, "")
	fileName = reAtMention.ReplaceAllString(fileName, "")
	fileName = reExtension.ReplaceAllString(fileName, "")

	// 2. Normaliza qualquer padrão de temporada/episódio para SXXEXX
	fileName = NormalizeEpisodeFormat(fileName)

	// 3. SE FOR SÉRATION (contém SXXEXX): limpa o título e junta puramente com o SXXEXX
	if matches := reSeasonEpisodePattern.FindStringIndex(fileName); len(matches) == 2 {
		rawTitle := fileName[:matches[0]]
		seasonEpisode := strings.ToUpper(fileName[matches[0]:matches[1]])

		// Limpa lixo do título (DSNP, C76, 1, H.264, Final, etc.)
		cleanTitle := reJunkTags.ReplaceAllString(rawTitle, "")
		cleanTitle = reYearWrapped.ReplaceAllString(cleanTitle, "")
		cleanTitle = reYearPlain.ReplaceAllString(cleanTitle, "")
		cleanTitle = reParensBrackets.ReplaceAllString(cleanTitle, "")

		// Remove números isolados residuais do título (como o "1" solto em DSNP.1)
		cleanTitle = regexp.MustCompile(`\b\d+\b`).ReplaceAllString(cleanTitle, "")

		// Substitui múltiplos espaços/pontos/traços por um único ponto
		cleanTitle = regexp.MustCompile(`[_\.\-\s]+`).ReplaceAllString(cleanTitle, ".")
		cleanTitle = strings.Trim(cleanTitle, ".")

		return cleanTitle + "." + seasonEpisode
	}

	// 4. SE FOR FILME: Mantém o fluxo normal
	qualityInfo := extractQualityInfo(fileName)
	fileName = reYearWrapped.ReplaceAllString(fileName, "")
	fileName = reYearPlain.ReplaceAllString(fileName, "")
	fileName = reResolution.ReplaceAllString(fileName, "")
	fileName = reParensBrackets.ReplaceAllString(fileName, "")
	fileName = reQualityCodecs.ReplaceAllString(fileName, "")
	fileName = reReleaseGroup.ReplaceAllString(fileName, "")
	fileName = reJunkTags.ReplaceAllString(fileName, "")

	fileName = regexp.MustCompile(`[_\.\-\s]+`).ReplaceAllString(fileName, ".")
	fileName = strings.Trim(fileName, ".")

	if qualityInfo != "" {
		fileName = fileName + "." + qualityInfo
	}

	return fileName
}

// extractQualityInfo extrai informações de qualidade, codec e áudio do nome do arquivo
// Mantém o case original (BRRip, H265, AAC) como escrito no arquivo
// Retorna uma string com: [qualidade].[codec].[audio]
func extractQualityInfo(fileName string) string {
	var parts []string

	// Extrai qualidade usando regex
	if matches := reQualityPattern.FindStringSubmatch(fileName); len(matches) > 1 {
		parts = append(parts, matches[1])
	}

	// Extrai codec usando regex
	if matches := reCodecPattern.FindStringSubmatch(fileName); len(matches) > 1 {
		parts = append(parts, matches[1])
	}

	// Extrai áudio usando regex
	if matches := reAudioPattern.FindStringSubmatch(fileName); len(matches) > 1 {
		parts = append(parts, matches[1])
	}

	return strings.Join(parts, ".")
}

// removeDuplicateEpisodes remove episódios/temporadas duplicados que aparecem no início do nome
// O episódio correto é aquele que vem DEPOIS dos traços " - - "
// Preserva a qualidade e codecs que estavam na primeira ocorrência
// Exemplo: "Bel_Air.S03E01...WEB_DL_H_264_C76 - - S03E10" -> "Bel_Air.S03E10.WEB-DL.H.264"
func removeDuplicateEpisodes(fileName string) string {
	// Encontra todos os padrões S##E## no arquivo
	matches := reSeasonEpisodePattern.FindAllStringIndex(fileName, -1)
	if len(matches) <= 1 {
		return fileName
	}

	// Se há múltiplos episódios, mantém o ÚLTIMO (que vem após " - - ")
	// A qualidade geralmente está na parte ENTRE episódios

	firstMatch := matches[0]
	lastMatch := matches[len(matches)-1]

	// Pega a parte antes do primeiro episódio (nome da série/filme)
	beforeFirstEpisode := fileName[:firstMatch[0]]

	// Remove underscores e pontos/traços no final de beforeFirstEpisode
	beforeFirstEpisode = strings.TrimRight(beforeFirstEpisode, "._- ")

	// Pega a parte entre o primeiro episódio e o último episódio (pode conter qualidade)
	betweenEpisodes := fileName[firstMatch[1]:lastMatch[0]]

	// Pega o último episódio correto
	lastEpisode := fileName[lastMatch[0]:lastMatch[1]]

	// Pega tudo depois do último episódio (qualidade, codec, etc)
	afterLastEpisode := fileName[lastMatch[1]:]

	// Extrai informações de qualidade/codec/áudio da parte entre os episódios (normalmente ali)
	qualityFromBetween := extractQualityInfoFromString(betweenEpisodes)

	// Reconstrói: nome + ponto + último_episódio + qualidade_dos_traços + após_episódio
	result := beforeFirstEpisode

	// Garante que haja um separador antes do episódio
	if len(result) > 0 && !strings.HasSuffix(result, ".") {
		result = result + "."
	}

	result = result + lastEpisode

	if qualityFromBetween != "" {
		result = result + "." + qualityFromBetween
	}

	result = result + afterLastEpisode

	// Remove pontos no início se houver
	result = strings.TrimPrefix(result, ".")

	return result
}

// extractQualityInfoFromString extrai qualidade de uma string específica
func extractQualityInfoFromString(s string) string {
	var parts []string

	// Extrai qualidade usando regex pré-compilada
	if matches := reQualityPattern.FindStringSubmatch(s); len(matches) > 1 {
		parts = append(parts, matches[1])
	}

	// Extrai codec usando regex pré-compilada
	if matches := reCodecPattern.FindStringSubmatch(s); len(matches) > 1 {
		parts = append(parts, matches[1])
	}

	// Extrai áudio usando regex pré-compilada
	if matches := reAudioPattern.FindStringSubmatch(s); len(matches) > 1 {
		parts = append(parts, matches[1])
	}

	return strings.Join(parts, ".")
}

// cleanNameForDownload remove emojis e caracteres problemáticos que impedem download direto
// Remove caracteres como :, <, >, ?, ", |, e emojis/símbolos Unicode
func cleanNameForDownload(text string) string {
	if text == "" {
		return text
	}

	// Remove emojis e símbolos Unicode
	text = strings.Map(func(r rune) rune {
		// Remove caracteres que são símbolos, emojis ou marcas Unicode
		if unicode.IsSymbol(r) || r > 0x1F000 || unicode.IsMark(r) {
			return -1 // Remove o caractere
		}
		return r
	}, text)

	// Remove caracteres problemáticos para downloads em Windows/Linux
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

	// Remove extensão
	nameWithoutExt := fileName
	re := regexp.MustCompile(`\.[a-zA-Z0-9]+$`)
	nameWithoutExt = re.ReplaceAllString(fileName, "")

	// Normaliza para lowercase para comparação
	lowerName := strings.ToLower(strings.TrimSpace(nameWithoutExt))

	// Lista de nomes genéricos
	genericNames := map[string]bool{
		"file":         true,
		"arquivo":      true,
		"document":     true,
		"documento":    true,
		"video":        true,
		"vídeo":        true,
		"audio":        true,
		"áudio":        true,
		"image":        true,
		"imagem":       true,
		"photo":        true,
		"foto":         true,
		"media":        true,
		"mídia":        true,
		"download":     true,
		"unnamed":      true,
		"untitled":     true,
		"noname":       true,
		"unknown":      true,
		"desconhecido": true,
		"sem nome":     true,
	}

	if genericNames[lowerName] {
		return true
	}

	// Verifica se é nome de site de download (padrões comuns)
	// Remove informações de episódio para pegar só o título base
	baseName := lowerName
	baseName = regexp.MustCompile(`(?i)\s*-?\s*s\d+e\d+.*$`).ReplaceAllString(baseName, "")
	baseName = regexp.MustCompile(`(?i)\s*-?\s*\d+x\d+.*$`).ReplaceAllString(baseName, "")
	baseName = strings.TrimSpace(baseName)

	// Padrões de nomes de sites
	sitePatterns := []string{
		"baixar.*mp4",
		"download.*mp4",
		"filmesmp4",
		"seriesmp4",
		"comandotorrents",
		"comandofilmes",
		"torrentsbr",
		"bludv",
		"mega.*filmes",
		"series.*online",
	}

	for _, pattern := range sitePatterns {
		matched, _ := regexp.MatchString("(?i)"+pattern, baseName)
		if matched {
			return true
		}
	}

	return false
}

// ExtractQualityFromDescription extrai informações de qualidade da descrição da mensagem
// Procura por padrões como: FHD, HD, SD, WEB-DL, Dublado, etc
// Retorna padrão: [Qualidade].[Fonte].[Áudio]
func ExtractQualityFromDescription(description string) string {
	if description == "" {
		return ""
	}

	var qualityParts []string

	// Padrões de qualidade/resolução (ordem importa - FHD tem prioridade sobre HD)
	qualityPatterns := []struct {
		pattern string
		label   string
	}{
		{`(?i)\bFHD\b`, "FHD"},
		{`(?i)\b(1080p|1080|Full.?HD)\b`, "FHD"},
		{`(?i)\b4K\b`, "4K"},
		{`(?i)\b(720p|720|HD)\b`, "HD"},
		{`(?i)\bSD\b`, "SD"},
	}

	// Padrões de fonte/tipo (ordem importa - prioridade)
	sourcePatterns := []struct {
		pattern string
		label   string
	}{
		{`(?i)WEB-?DL`, "WEB-DL"},
		{`(?i)BluRay`, "BluRay"},
		{`(?i)WEB-?RIP`, "WEBRip"},
		{`(?i)HDRip`, "HDRip"},
	}

	// Padrões de áudio/idioma (ordem importa - extrai todos, não só primeiro)
	audioPatterns := []struct {
		pattern string
		label   string
	}{
		{`(?i)Dublado|DUB`, "Dub"},
		{`(?i)Legendado|LEG`, "Leg"},
		{`PT-BR|🇧🇷|(?i)Brasil`, "PT"},
		{`(?i)Portugu[eê]s`, "PT"},
	}

	// Extrai qualidade/resolução (primeira match tem prioridade)
	for _, q := range qualityPatterns {
		re := regexp.MustCompile(q.pattern)
		if re.MatchString(description) {
			qualityParts = append(qualityParts, q.label)
			break
		}
	}

	// Extrai fonte (primeira match tem prioridade)
	for _, s := range sourcePatterns {
		re := regexp.MustCompile(s.pattern)
		if re.MatchString(description) {
			qualityParts = append(qualityParts, s.label)
			break
		}
	}

	// Extrai áudio/idioma (pode ter múltiplos, mas em ordem de prioridade)
	// Verifica todos os padrões em ordem e adiciona o primeiro que achar
	audioFound := make(map[string]bool)
	for _, a := range audioPatterns {
		re := regexp.MustCompile(a.pattern)
		if re.MatchString(description) && !audioFound[a.label] {
			qualityParts = append(qualityParts, a.label)
			audioFound[a.label] = true
		}
	}

	// Retorna partes juntas com ponto
	if len(qualityParts) > 0 {
		return strings.Join(qualityParts, ".")
	}

	return ""
}
