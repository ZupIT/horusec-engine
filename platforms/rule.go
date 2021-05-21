package platforms

import (
	"github.com/antchfx/xpath"

	engine "github.com/ZupIT/horusec-engine"
)

type MatchType int

const (
	RegularMatch MatchType = iota
	NotMatch
)

type StructuredDataRule struct {
	engine.Metadata
	Type        MatchType
	Expressions []*xpath.Expr
}

func (rule StructuredDataRule) IsFor(unitType engine.UnitType) bool {
	return engine.StructuredDataUnit == unitType
}

func NewStructuredDataRule(matchType MatchType, queryStrings []string) StructuredDataRule {
	var exprs []*xpath.Expr
	for _, query := range queryStrings {
		exprs = append(exprs, xpath.MustCompile(query))
	}

	return StructuredDataRule{
		Type:        matchType,
		Expressions: exprs,
	}
}
