package engine


type Unit interface {
	Type() UnitType
	Eval(Rule) []Finding
}

type Finding struct {
	ID             string
	SourceLocation Location
}

func Run(document []Unit, rules []Rule) IOutput {
	return NewOutput(run(document, rules))
}

func run(document []Unit, rules []Rule) (documentFindings []Finding) {
	if len(document) < 1 || len(rules) < 1 {
		return []Finding{}
	}

	numberOfUnits := (len(document)) * (len(rules))

	documentFindingsChannel := make(chan []Finding, numberOfUnits)

	for _, documentUnit := range document {
		go execRulesInDocumentUnit(rules, documentUnit, documentFindingsChannel)
	}

	for i := 1; i <= numberOfUnits; i++ {
		unitFindings := <-documentFindingsChannel
		documentFindings = append(documentFindings, unitFindings...)
	}

	close(documentFindingsChannel)

	return documentFindings
}

func execRulesInDocumentUnit(rules []Rule, documentUnit Unit, findings chan<- []Finding) {
	for _, rule := range rules {
		if rule.IsFor(documentUnit.Type()) {
			go func() {
				ruleFindings := documentUnit.Eval(rule)
				findings <- ruleFindings
			}()
		}
	}
}
