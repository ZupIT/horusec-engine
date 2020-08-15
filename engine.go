package engine

import (
	"github.com/ZupIT/horus-engine/fs"
)

type Unit interface {
	Eval(rule Rule) []Finding
}

type Location struct {
	Filename string
	Line     string
	Column   string
}

type Finding struct {
	ID             string
	SourceLocation Location
}

func ExecRulesInDocumentUnit(rules []Rule, documentUnit Unit, findings chan<- []Finding) {
	for _, rule := range rules {
		go func() {
			ruleFindings := documentUnit.Eval(rule)
			findings <- ruleFindings
		}()
	}
}

func Run(document []Unit, rules []Rule) (documentFindings []Finding) {
	numberOfUnits := len(document)

	documentFindingsChannel := make(chan []Finding, numberOfUnits)

	for _, documentUnit := range document {
		go ExecRulesInDocumentUnit(rules, documentUnit, documentFindingsChannel)
	}

	for i = 0; i <= numberOfUnits; i++ {
		unitFindings <- documentFindingsChannel
		documentFindings = append(documentFindings, unitFindings...)

	}
}
