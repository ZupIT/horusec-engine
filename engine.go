package engine

type Unit interface {
	Type() UnitType
	Eval(rule Rule) []Finding
}

type Location struct {
	Filename string
	Line     int
	Column   int
}

type Finding struct {
	ID             string
	SourceLocation Location
}

func ExecRulesInDocumentUnit(rules []Rule, documentUnit Unit, findings chan<- []Finding) {
	for _, rule := range rules {
		if rule.IsFor() == documentUnit.Type() {
			go func() {
				ruleFindings := documentUnit.Eval(rule)
				findings <- ruleFindings
			}()
		}
	}
}

func Run(document []Unit, rules []Rule) (documentFindings []Finding) {
	numberOfUnits := len(document) - 1

	documentFindingsChannel := make(chan []Finding, numberOfUnits)

	for _, documentUnit := range document {
		go ExecRulesInDocumentUnit(rules, documentUnit, documentFindingsChannel)
	}

	for i := 0; i <= numberOfUnits; i++ {
		unitFindings := <-documentFindingsChannel
		documentFindings = append(documentFindings, unitFindings...)
	}

	close(documentFindingsChannel)

	return documentFindings
}
