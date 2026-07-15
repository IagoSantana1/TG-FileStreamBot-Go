package utils

import (
	"fmt"
	"regexp"
	"strconv"
)

// Regex compiladas uma única vez para máxima performance
var (
	// 1. Unifica 0x0, T00E00 e S00E00 em uma única expressão para normalização rápida
	reUnifiedEpisode = regexp.MustCompile(`(?i)\b(?:(\d{1,2})[xX](\d{1,2})|[ST](\d{1,2})E(\d{1,2}))\b`)

	// 2. Unifica padrões de verificação de temporada (S01E01 ou 01x01)
	reSeasonPatternUnified = regexp.MustCompile(`(?i)\b(?:S\d{1,2}E\d{1,3}|\d{1,2}[xX]\d{1,2})\b`)
	rePtSeasonPattern      = regexp.MustCompile(`(?i)temporada[\s._-]*\d{1,2}.*epis[óo]dio[\s._-]*\d{1,3}`)

	// 3. Unifica TODOS os formatos de episódio isolado (E05, EP05, Episódio 5, #Episódio 05, etc.) em uma única busca
	reUnifiedEpisodeOnly = regexp.MustCompile(`(?i)(#\s*(?:episodio|episódio|ep)|\b(?:E[Pp]?|episodio|episódio|capitulo|capítulo|ep))\s*\.?\s*0*(\d{1,3})\b`)

	reYearPattern = regexp.MustCompile(`\b(19|20)\d{2}\b`)
	reSeasonNum   = regexp.MustCompile(`(?i)\b(?:season|temporada)[\s._-]*0*(\d{1,2})\b`)
	reShortSeason = regexp.MustCompile(`(?i)\bT\s*0*(\d{1,2})\b`)
)

// NormalizeEpisodeFormat normaliza diferentes formatos de episódio para S00E00
func NormalizeEpisodeFormat(fileName string) string {
	return normalizeEpisodeFormat(fileName)
}

// normalizeEpisodeFormat realiza a conversão em uma única passada de alta performance
func normalizeEpisodeFormat(fileName string) string {
	return reUnifiedEpisode.ReplaceAllStringFunc(fileName, func(match string) string {
		parts := reUnifiedEpisode.FindStringSubmatch(match)
		if len(parts) == 5 {
			// Se casou com o formato '1x02' (primeira parte do OU da regex)
			if parts[1] != "" {
				season, _ := strconv.Atoi(parts[1])
				episode, _ := strconv.Atoi(parts[2])
				return fmt.Sprintf("S%02dE%02d", season, episode)
			}
			// Se casou com o formato 'S01E02' ou 'T01E02' (segunda parte do OU da regex)
			if parts[3] != "" {
				season, _ := strconv.Atoi(parts[3])
				episode, _ := strconv.Atoi(parts[4])
				return fmt.Sprintf("S%02dE%02d", season, episode)
			}
		}
		return match
	})
}

// HasSeasonPattern verifica se o nome já contém padrão de temporada+episódio
func HasSeasonPattern(fileName string) bool {
	return reSeasonPatternUnified.MatchString(fileName) || rePtSeasonPattern.MatchString(fileName)
}

// HasEpisodeOnlyPattern verifica se há padrão de episódio sem temporada clara
func HasEpisodeOnlyPattern(fileName string) bool {
	if HasSeasonPattern(fileName) {
		return false
	}
	return reUnifiedEpisodeOnly.MatchString(fileName)
}

// IsLikelyMovie tenta detectar se é filme (sem padrões de episódio e com ano)
func IsLikelyMovie(fileName string) bool {
	if HasSeasonPattern(fileName) || HasEpisodeOnlyPattern(fileName) {
		return false
	}
	return reYearPattern.MatchString(fileName)
}

// ExtractSeasonNumber tenta extrair o número da temporada de "Season X" ou "season X"
func ExtractSeasonNumber(fileName string) (int, bool) {
	if matches := reSeasonNum.FindStringSubmatch(fileName); len(matches) == 2 {
		season, err := strconv.Atoi(matches[1])
		if err == nil && season > 0 {
			return season, true
		}
	}

	if matches := reShortSeason.FindStringSubmatch(fileName); len(matches) == 2 {
		season, err := strconv.Atoi(matches[1])
		if err == nil && season > 0 {
			return season, true
		}
	}

	return 0, false
}

// InjectSeasonIntoEpisode injeta temporada em nomes com episódio de forma otimizada (sem loops)
func InjectSeasonIntoEpisode(fileName string, season int) string {
	if HasSeasonPattern(fileName) {
		return normalizeEpisodeFormat(fileName)
	}

	// Faz a substituição de qualquer formato de episódio isolado em uma única passada!
	updated := reUnifiedEpisodeOnly.ReplaceAllStringFunc(fileName, func(match string) string {
		parts := reUnifiedEpisodeOnly.FindStringSubmatch(match)
		if len(parts) == 3 {
			// O segundo grupo de captura contêm o número do episódio isolado
			ep, _ := strconv.Atoi(parts[2])
			return fmt.Sprintf("S%02dE%02d", season, ep)
		}
		return match
	})

	return normalizeEpisodeFormat(updated)
}
