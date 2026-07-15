package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// Regex pré-compiladas para melhor performance
var (
	reSiga                 = regexp.MustCompile(`(?i)\bsiga\b.*`)
	reAtMention            = regexp.MustCompile(`@\S+`)
	reExtension            = regexp.MustCompile(`(?i)\.(mp4|mkv|avi)$`)
	reYearWrapped          = regexp.MustCompile(`\(\d{4}\)|\[\d{4}\]`)
	reYearPlain            = regexp.MustCompile(`\b\d{4}\b`)
	reResolution           = regexp.MustCompile(`(?i)\b\d{3,4}p\b`)
	reParensBrackets       = regexp.MustCompile(`[\(\)\[\]]`)
	reQualityCodecs        = regexp.MustCompile(`(?i)\b(web[-\s]?dl|webdl|web[-\s]?rip|webrip|blu[-\s]?ray|bluray|hd[-\s]?rip|hdrip|dvd[-\s]?rip|dvdrip|hdtv|brip|brrip|4k|uhd|x264|x265|h\.?264|h\.?265|hevc|avc|ddp?\d*\.?\d*|aac\d*\.?\d*|ac3|dts[-\s]?hd|dts|atmos|truehd)\b`)
	reReleaseGroup         = regexp.MustCompile(`-[A-Za-z]+`)
	reDash                 = regexp.MustCompile(`\s*-\s*`)
	reMultipleDots         = regexp.MustCompile(`\.{2,}`)
	reSeasonEpisodeCase    = regexp.MustCompile(`(?i)\bs(\d+)e(\d+)\b`)
	reQualityPattern       = regexp.MustCompile(`(?i)\b(BRRip|BRRIP|brip|BRip|WEB-DL|WEB-RIP|BluRay|Blu-Ray|HDRip|HD-RIP|DVDRip|DVD-RIP|HDTV|WebRip|WEBRip)\b`)
	reCodecPattern         = regexp.MustCompile(`(?i)\b(x265|x264|H\.?265|H\.?264|HEVC|AVC|HEVC)\b`)
	reAudioPattern         = regexp.MustCompile(`(?i)\b(AAC|AC3|DDP5\.1|DDP\s5\.1|DTS|DTS-HD|ATMOS|TrueHD)\b`)
	reKnownSuffix          = regexp.MustCompile(`(?i)^(hdrip|hdtv|brrip|bdrip|webrip|web[-_]?dl|bluray|brip|uhd|4k|fhd|hd|sd|x264|x265|h264|h265|hevc|avc|aac\d*|ac3|ddp\d*(?:\.\d+)?|dts[-_]?hd|dts|atmos|truehd|\d{3,4}p|s\d{1,2}e\d{1,3}|\d{1,2}[xX]\d{1,2})$`)
	reSeasonEpisodePattern = regexp.MustCompile(`(?i)s\d{1,2}e\d{1,3}`)
)

func ProcessStrmFileName(fileName string) string {
	return ProcessStrmFileNameWithQuality(fileName, 0)
}

func ProcessStrmFileNameWithQuality(fileName string, fileSize int64) string {
	if fileName == "" {
		return fileName
	}

	// Remove espaços no início e fim
	fileName = strings.TrimSpace(fileName)

	// Se tiver quebras de linha, usa apenas a primeira linha (antes de quebras)
	lines := strings.Split(fileName, "\n")
	fileName = strings.TrimSpace(lines[0])

	// Remove "Siga:" e tudo que vem após (case insensitive)
	fileName = reSiga.ReplaceAllString(fileName, "")

	// Remove menções com @ (ex: @ultrafilmesbr)
	fileName = reAtMention.ReplaceAllString(fileName, "")

	// Remove espaços extras novamente após remoções
	fileName = strings.TrimSpace(fileName)

	// Remove ponto inicial se existir
	fileName = strings.TrimPrefix(fileName, ".")

	// Remove extensões de mídia (.mp4, .mkv, .avi, .mov, .flv, .wmv, .webm, etc - case insensitive)
	fileName = reExtension.ReplaceAllString(fileName, "")

	// *** IMPORTANTE: Extrai qualidade ANTES de remover duplicatas ***
	// Isso garante que capturemos a qualidade de ANTES de " - - "
	qualityInfo := extractQualityInfo(fileName)

	/// Remove datas entre parênteses: (2011), (2023), etc.
	fileName = reYearWrapped.ReplaceAllString(fileName, "")

	// Remove anos soltos (4 dígitos)
	fileName = reYearPlain.ReplaceAllString(fileName, "")

	// Remove resoluções de vídeo: 720p, 1080p, 2160p, etc.
	fileName = reResolution.ReplaceAllString(fileName, "")

	// Remove caracteres entre parênteses e colchetes
	fileName = reParensBrackets.ReplaceAllString(fileName, "")

	// Remove qualidade, codecs e áudio do meio (serão adicionados no final)
	fileName = reQualityCodecs.ReplaceAllString(fileName, "")

	// Remove grupos de release (ex: -KANE, -GROUP)
	fileName = reReleaseGroup.ReplaceAllString(fileName, "")

	// Remove traços isolados
	fileName = reDash.ReplaceAllString(fileName, ".")

	// Remove espaços extras
	fileName = strings.TrimSpace(fileName)

	// Substitui espaços por pontos
	fileName = strings.ReplaceAll(fileName, " ", ".")

	// Remove pontos múltiplos consecutivos
	fileName = reMultipleDots.ReplaceAllString(fileName, ".")

	// Remove pontos no início e fim
	fileName = strings.Trim(fileName, ".")

	// Remove episódios/temporadas duplicados no final do nome
	// Também preserva a qualidade extraída anteriormente
	fileName = removeDuplicateEpisodes(fileName)

	// Garante que S e E de temporada/episódio estejam em MAIÚSCULO (s01e01 -> S01E01)
	// Normaliza diferentes formatos de episódio para S00E00 (ex: 1x01, T01E01 -> S01E01)
	fileName = NormalizeEpisodeFormat(fileName)

	// Garante que S e E de temporada/episódio estejam em MAIÚSCULO (s01e01 -> S01E01)
	fileName = reSeasonEpisodeCase.ReplaceAllStringFunc(fileName, func(match string) string {
		return strings.ToUpper(match)
	})

	// Adiciona informações de qualidade no final (se existirem)
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
