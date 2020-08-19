package text

import (
	"regexp"

	"github.com/ZupIT/horus-engine"
)

type TextRule struct {
	ID          string
	Expressions []*regexp.Regexp
	MatchType   engine.MatchType
}

func (rule TextRule) IsFor() engine.UnitType {
	return engine.ProgramTextUnit
}

func (rule TextRule) Type() engine.MatchType {
	return rule.MatchType
}
