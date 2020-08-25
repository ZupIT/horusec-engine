package engine

type Unit interface {
	Type() UnitType
	Eval(Rule) []Finding
}

type Finding struct {
	ID             string
	SourceLocation Location
}

type Report struct {
	ID             string   // Comes from Advisory::GetID/0
	Name           string   // Comes from Advisory::GetName/0
	Description    string   // Comes from Advisory::GetDescription/0
	SourceLocation Location // Comes from the Finding
}

func BuildReport(findings []Finding, advisories []Advisory) (programReport []Report) {
	for _, advisory := range advisories {
		for _, finding := range findings {
			if finding.ID == advisory.GetID() {
				report := Report{
					ID:             advisory.GetID(),
					Name:           advisory.GetName(),
					Description:    advisory.GetDescription(),
					SourceLocation: finding.SourceLocation,
				}

				programReport = append(programReport, report)
			}
		}
	}

	return programReport
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

func Run(document []Unit, rules []Rule) (documentFindings []Finding) {
	if len(document) < 1 || len(rules) < 1 {
		return []Finding{}
	}

	numberOfUnits := ((len(document)) * (len(rules)))

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
