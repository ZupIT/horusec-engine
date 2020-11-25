package android

import (
	"io/ioutil"
	"testing"

	engine "github.com/ZupIT/horusec-engine"
	"github.com/ZupIT/horusec-engine/platforms"
)

func TestLoadingAManifestUnitWithValidManifestShouldWork(t *testing.T) {
	androidManifest, err := ioutil.ReadFile("AndroidManifest.xml")

	if err != nil {
		t.Fatal(err)
	}

	_, err = NewManifestUnit(androidManifest)

	if err != nil {
		t.Fatal(err)
	}
}

func TestMatchRegularRuleWithValidManifestShouldWork(t *testing.T) {
	androidManifest, err := ioutil.ReadFile("AndroidManifest.xml")

	if err != nil {
		t.Fatal(err)
	}

	manifestUnit, err := NewManifestUnit(androidManifest)

	if err != nil {
		t.Fatal(err)
	}

	exportedRule := platforms.NewStructuredDataRule(platforms.RegularMatch, []string{`//manifest//application//activity[@android:exported='true']`})

	findings := engine.Run([]engine.Unit{manifestUnit}, []engine.Rule{exportedRule})

	if len(findings) <= 0 {
		t.Fatal("Should have found something")
	}
}

func TestMatchNotRuleWithValidManifestShouldWork(t *testing.T) {
	androidManifest, err := ioutil.ReadFile("AndroidManifest.xml")

	if err != nil {
		t.Fatal(err)
	}

	manifestUnit, err := NewManifestUnit(androidManifest)

	if err != nil {
		t.Fatal(err)
	}

	exportedRule := platforms.NewStructuredDataRule(platforms.NotMatch, []string{`//manifest//application[@usesCleartextTraffic='true']`})
	exportedRule.Description = "Congratulations! You're not using the usesCleattextTraffic property on your applications!"

	findings := engine.Run([]engine.Unit{manifestUnit}, []engine.Rule{exportedRule})

	if len(findings) <= 0 {
		t.Fatal("Should have found something")
	}
}

func TestMatchNotRuleWithValidManifestShouldWorkFindingAnIssue(t *testing.T) {
	androidManifest, err := ioutil.ReadFile("AndroidManifest.2.xml")

	if err != nil {
		t.Fatal(err)
	}

	manifestUnit, err := NewManifestUnit(androidManifest)

	if err != nil {
		t.Fatal(err)
	}

	exportedRule := platforms.NewStructuredDataRule(platforms.NotMatch, []string{`//manifest//application[@android:usesCleartextTraffic='true']`})
	exportedRule.Description = "Congratulations! You're not using the usesCleattextTraffic property on your applications!"

	findings := engine.Run([]engine.Unit{manifestUnit}, []engine.Rule{exportedRule})

	if len(findings) > 0 {
		t.Fatal("Should not have found something")
	}
}

func TestCustomXPathExpressionsHandlingWithValidManifestShouldWork(t *testing.T) {
	androidManifest, err := ioutil.ReadFile("AndroidManifest.xml")

	if err != nil {
		t.Fatal(err)
	}

	manifestUnit, err := NewManifestUnit(androidManifest)

	if err != nil {
		t.Fatal(err)
	}

	exportedRule := platforms.NewStructuredDataRule(platforms.RegularMatch, []string{`//manifest//application//activity[@android:name[
		contains(
			translate(., 'ABCDEFGHIJKLMNOPQRSTUVWXYZ','abcdefghijklmnopqrstuvwxyz'),
			'smali')
		]]`})

	findings := engine.Run([]engine.Unit{manifestUnit}, []engine.Rule{exportedRule})

	if len(findings) <= 0 {
		t.Fatal("Should have found something")
	}
}
