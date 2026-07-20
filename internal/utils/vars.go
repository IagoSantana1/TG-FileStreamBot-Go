package utils

import "regexp"

var (
	// 🚀 Captura "Temporada 01 episódio 08" sem depender do \b (funciona com _, ., - ou espaços)
	rePtSeasonEpisodeCapture = regexp.MustCompile(`(?i)(?:^|[\s._-])temporada[\s._-]*0*(\d{1,2})[\s._-]*epis[óo]dio[\s._-]*0*(\d{1,3})(?:$|[\s._-])`)

	// 🚀 Tags de plataformas, codecs, resoluções e palavras de desfecho para remoção
	reJunkTags = regexp.MustCompile(`(?i)\b(DSNP|NF|AMZN|HMAX|ATVP|DSNY|PCOK|iT|CWA|C76|WEB-?DL|WEBRip|WEB|DL|BluRay|BRRip|HDRip|H\.?264|H\.?265|x264|x265|HEVC|AVC|AAC\d*|AC3|DDP?\d*[\._\s]?\d*|Atmos|TrueHD|DTS|2160p|1080p|720p|480p|4k|2k|Final|Completo|Dublado|DUAL|Legendado|Multi)\b`)

	// Seus padrões preexistentes mantidos abaixo:
	reUnifiedEpisode       = regexp.MustCompile(`(?i)\b(?:(\d{1,2})[xX](\d{1,2})|[ST](\d{1,2})E(\d{1,2}))\b`)
	reSeasonPatternUnified = regexp.MustCompile(`(?i)\b(?:S\d{1,2}E\d{1,3}|\d{1,2}[xX]\d{1,2})\b`)
	rePtSeasonPattern      = regexp.MustCompile(`(?i)temporada[\s._-]*\d{1,2}.*epis[óo]dio[\s._-]*\d{1,3}`)
	reUnifiedEpisodeOnly   = regexp.MustCompile(`(?i)(#\s*(?:episodio|episódio|ep)|\b(?:E[Pp]?|episodio|episódio|capitulo|capítulo|ep))\s*\.?\s*0*(\d{1,3})\b`)
	reYearPattern          = regexp.MustCompile(`\b(19|20)\d{2}\b`)
	reSeasonNum            = regexp.MustCompile(`(?i)\b(?:season|temporada)[\s._-]*0*(\d{1,2})\b`)
	reShortSeason          = regexp.MustCompile(`(?i)\bT\s*0*(\d{1,2})\b`)
	reSiga                 = regexp.MustCompile(`(?i)\bsiga\b.*`)
	reAtMention            = regexp.MustCompile(`@\S+`)
	reExtension            = regexp.MustCompile(`(?i)\.(mp4|mkv|avi|strm)$`)
	reYearWrapped          = regexp.MustCompile(`\(\d{4}\)|\[\d{4}\]`)
	reYearPlain            = regexp.MustCompile(`\b\d{4}\b`)
	reResolution           = regexp.MustCompile(`(?i)\b\d{3,4}p\b`)
	reParensBrackets       = regexp.MustCompile(`[\(\)\[\]]`)
	reQualityCodecs        = regexp.MustCompile(`(?i)\b(web[-\s]?dl|webdl|web[-\s]?rip|webrip|blu[-\s]?ray|bluray|hd[-\s]?rip|hdrip|dvd[-\s]?rip|dvdrip|hdtv|brip|brrip|4k|uhd|x264|x265|h\.?264|h\.?265|hevc|avc|ddp?\d*\.?\d*|aac\d*\.?\d*|ac3|dts[-\s]?hd|dts|atmos|truehd)\b`)
	reReleaseGroup         = regexp.MustCompile(`-[A-Za-z0-9]+`)
	reDash                 = regexp.MustCompile(`\s*-\s*`)
	reMultipleDots         = regexp.MustCompile(`\.{2,}`)
	reSeasonEpisodeCase    = regexp.MustCompile(`(?i)\bs(\d+)e(\d+)\b`)
	reQualityPattern       = regexp.MustCompile(`(?i)\b(BRRip|BRRIP|brip|BRip|WEB-DL|WEB-RIP|BluRay|Blu-Ray|HDRip|HD-RIP|DVDRip|DVD-RIP|HDTV|WebRip|WEBRip)\b`)
	reCodecPattern         = regexp.MustCompile(`(?i)\b(x265|x264|H\.?265|H\.?264|HEVC|AVC)\b`)
	reAudioPattern         = regexp.MustCompile(`(?i)\b(AAC|AC3|DDP5\.1|DDP\s5\.1|DTS|DTS-HD|ATMOS|TrueHD)\b`)
	reSeasonEpisodePattern = regexp.MustCompile(`(?i)s\d{1,2}e\d{1,3}`)
)
