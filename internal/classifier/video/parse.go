package video

import (
	"github.com/bitmagnet-io/bitmagnet/internal/classifier"
	"github.com/bitmagnet-io/bitmagnet/internal/model"
	"github.com/bitmagnet-io/bitmagnet/internal/regex"
	"github.com/hedhyw/rex/pkg/dialect"
	"github.com/hedhyw/rex/pkg/rex"
	"strconv"
)

var titleTokens = []dialect.Token{
	rex.Group.Define(
		rex.Group.Composite(
			rex.Group.NonCaptured(
				regex.AnyWordChar().Repeat().OneOrMore(),
				rex.Group.NonCaptured(
					rex.Chars.Single('-'), regex.AnyWordChar().Repeat().OneOrMore(),
				).Repeat().ZeroOrMore(),
			),
			regex.AnyNonWordChar().Repeat().OneOrMore(),
		).NonCaptured().Repeat().OneOrMore(),
		rex.Group.Composite(
			regex.AnyNonWordChar().Repeat().OneOrMore(),
			rex.Chars.End(),
		).NonCaptured(),
	),
}

var titleRegex = rex.New(
	rex.Chars.Begin(),
	rex.Group.NonCaptured(titleTokens...),
).MustCompile()

var yearTokens = []dialect.Token{
	rex.Group.NonCaptured(rex.Common.NotClass(rex.Chars.WordCharacter()).Repeat().ZeroOrMore()),
	rex.Group.Define(
		rex.Group.Composite(
			rex.Common.Text("18"), rex.Common.Text("19"), rex.Common.Text("20"),
		).NonCaptured(),
		rex.Chars.Digits().Repeat().Exactly(2),
	),
	rex.Group.Composite(
		rex.Common.NotClass(rex.Chars.WordCharacter()),
		rex.Chars.End(),
	).NonCaptured(),
}

var titleYearRegex = rex.New(
	rex.Chars.Begin(),
	rex.Group.NonCaptured(rex.Group.NonCaptured(titleTokens...), rex.Group.NonCaptured(yearTokens...)),
).MustCompile()

var titleEpisodesRegex = rex.New(
	rex.Chars.Begin(),
	rex.Group.NonCaptured(
		rex.Group.NonCaptured(titleTokens...),
		model.EpisodesToken,
	),
).MustCompile()

var multiRegex = regex.NewRegexFromNames("multi")

var separatorToken = rex.Chars.Runes(" ._")

var titlePartRegex = rex.New(
	separatorToken.Repeat().ZeroOrOne(),
	rex.Group.Define(regex.WordToken()),
	separatorToken.Repeat().ZeroOrOne(),
).MustCompile()

var trimTitleRegex = rex.New(
	rex.Chars.Begin(),
	rex.Group.NonCaptured(
		rex.Chars.Single('['),
		rex.Common.NotClass(rex.Chars.Single(']')).Repeat().OneOrMore(),
		rex.Chars.Single(']'),
	).Repeat().ZeroOrOne(),
	regex.AnyNonWordChar().Repeat().ZeroOrMore(),
	rex.Group.Define(
		regex.WordToken(),
		rex.Group.NonCaptured(
			rex.Chars.Any(),
			regex.WordToken(),
		).Repeat().ZeroOrMore(),
	),
	regex.AnyNonWordChar().Repeat().ZeroOrMore(),
	rex.Chars.End(),
).MustCompile()

func cleanTitle(title string) string {
	title = titlePartRegex.ReplaceAllStringFunc(title, func(s string) string {
		partMatch := titlePartRegex.FindStringSubmatch(s)
		if partMatch == nil {
			return ""
		}
		return partMatch[1] + " "
	})
	title = trimTitleRegex.ReplaceAllString(title, "$1")
	return title
}

func parseTitleYear(input string) (string, model.Year, string, error) {
	if match := titleYearRegex.FindStringSubmatch(input); match != nil {
		yearMatch, _ := strconv.ParseUint(match[2], 10, 16)
		title := cleanTitle(match[1])
		if title != "" {
			return title, model.Year(yearMatch), input[len(match[0]):], nil
		}
	}
	return "", 0, "", classifier.ErrNoMatch
}

func parseTitle(input string) (string, string, error) {
	if match := titleRegex.FindStringSubmatch(input); match != nil {
		title := cleanTitle(match[1])
		if title != "" {
			return title, input[len(match[0]):], nil
		}
	}
	return "", "", classifier.ErrNoMatch
}

func parseTitleYearEpisodes(input string) (string, model.Year, model.Episodes, string, error) {
	if match := titleEpisodesRegex.FindStringSubmatch(input); match != nil {
		title := match[1]
		year := model.Year(0)
		if t, y, _, err := parseTitleYear(title); err == nil {
			title = t
			year = y
		} else {
			title = cleanTitle(title)
		}
		episodes := model.EpisodesMatchToEpisodes(match[2:])
		return title, year, episodes, input[len(match[0]):], nil
	}
	return "", 0, nil, "", classifier.ErrNoMatch
}

func ParseTitleYearEpisodes(contentType model.NullContentType, input string) (string, model.Year, model.Episodes, string, error) {
	if !contentType.Valid || contentType.ContentType == model.ContentTypeTvShow {
		if title, year, episodes, rest, err := parseTitleYearEpisodes(input); err == nil {
			return title, year, episodes, rest, nil
		}
	}
	if title, year, rest, err := parseTitleYear(input); err == nil {
		return title, year, nil, rest, nil
	}
	if title, rest, err := parseTitle(input); err == nil {
		return title, 0, nil, rest, nil
	}
	return "", 0, nil, "", classifier.ErrNoMatch
}

func ParseContent(hintCt model.NullContentType, input string) (model.ContentType, string, model.Year, classifier.ContentAttributes, error) {
	title, year, episodes, rest, err := ParseTitleYearEpisodes(hintCt, input)
	if err != nil {
		return "", "", 0, classifier.ContentAttributes{}, err
	}
	var ct model.ContentType
	if hintCt.Valid {
		ct = hintCt.ContentType
	} else if len(episodes) > 0 {
		ct = model.ContentTypeTvShow
	} else {
		ct = model.ContentTypeMovie
	}
	if ct != model.ContentTypeTvShow {
		episodes = nil
	}
	vc, rg := model.InferVideoCodecAndReleaseGroup(rest)
	return ct, title, year, classifier.ContentAttributes{
		Episodes:        episodes,
		Languages:       model.InferLanguages(rest),
		LanguageMulti:   multiRegex.MatchString(rest),
		VideoResolution: model.InferVideoResolution(rest),
		VideoSource:     model.InferVideoSource(rest),
		VideoCodec:      vc,
		Video3d:         model.InferVideo3d(rest),
		VideoModifier:   model.InferVideoModifier(rest),
		ReleaseGroup:    rg,
	}, nil
}
