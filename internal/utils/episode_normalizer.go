package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// NormalizeEpisodeFormat normaliza diferentes formatos de episódio para S00E00
func NormalizeEpisodeFormat(fileName string) string {
	return normalizeEpisodeFormat(fileName)
}

// normalizeEpisodeFormat realiza a conversão em uma única passada de alta performance
func normalizeEpisodeFormat(fileName string) string {
	// Normaliza sublinhados para espaços para evitar que quebrem a busca de borda de palavra
	clean := strings.ReplaceAll(fileName, "_", " ")

	// 1. Converte "Temporada 01 episódio 08" -> "S01E08"
	if matches := rePtSeasonEpisodeCapture.FindStringSubmatch(clean); len(matches) >= 3 {
		season, _ := strconv.Atoi(matches[1])
		episode, _ := strconv.Atoi(matches[2])
		seStr := fmt.Sprintf("S%02dE%02d", season, episode)
		clean = rePtSeasonEpisodeCapture.ReplaceAllString(clean, " "+seStr+" ")
	}

	// 2. Converte 1x08 ou T01E08 -> S01E08
	clean = reUnifiedEpisode.ReplaceAllStringFunc(clean, func(match string) string {
		parts := reUnifiedEpisode.FindStringSubmatch(match)
		if len(parts) == 5 {
			if parts[1] != "" {
				s, _ := strconv.Atoi(parts[1])
				e, _ := strconv.Atoi(parts[2])
				return fmt.Sprintf("S%02dE%02d", s, e)
			}
			if parts[3] != "" {
				s, _ := strconv.Atoi(parts[3])
				e, _ := strconv.Atoi(parts[4])
				return fmt.Sprintf("S%02dE%02d", s, e)
			}
		}
		return match
	})

	return clean
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

// 🚀 ADICIONADO: Distribui uma lista de episódios sequencialmente para o lote de mídias
func DistributeEpisodes(totalFiles int, inputEps []int) []int {
	if len(inputEps) == 0 {
		inputEps = []int{1}
	}
	result := make([]int, totalFiles)
	startEp := inputEps[0]
	for i := 0; i < totalFiles; i++ {
		if i < len(inputEps) {
			result[i] = inputEps[i]
		} else {
			result[i] = startEp + i
		}
	}
	return result
}
