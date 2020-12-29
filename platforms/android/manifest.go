package android

import (
	"bytes"

	"github.com/antchfx/xmlquery"

	engine "github.com/ZupIT/horusec-engine"
	"github.com/ZupIT/horusec-engine/platforms"
)

// Permission holds data about all the declared permissions in an AndroidManifest.xml file
type Permission struct {
	Name     string `xml:"name,attr"`
	Required string `xml:"required,attr"`
}

// SDKInfo holds information about the target and compilation SDK of the given App
type SDKInfo struct {
	MinimumSDKVersion string `xml:"minSdkVersion,attr"`
	TargetSDKVersion  string `xml:"targetSdkVersion,attr"`
	MaximumSDKVersion string `xml:"maxSdkVersion,attr"`
}

// IntentAction holds data about the declared Actions that can be performed in given Activity.
type IntentAction struct {
	Name string `xml:"name,attr"`
}

// IntentCategory is the Activity's category.
type IntentCategory struct {
	Name string `xml:"name,attr"`
}

// IntentFilter holds imformational data about the `intention-filter` tag for the given Activity.
type IntentFilter struct {
	Categories IntentCategory `xml:"category"`
	Actions    []IntentAction `xml:"action"`
}

// Activity represents an Activity entry in the manifest file
type Activity struct {
	Name         string       `xml:"name,attr"`
	IntentFilter IntentFilter `xml:"intent-filter"`
}

// BroadcastReceiver represents a broadcast receiver entry in the manifest file
type BroadcastReceiver struct {
	Name       string `xml:"name,attr"`
	Enabled    string `xml:"enabled,attr"`
	IsExported string `xml:"exported,attr"`
	Permission string `xml:"permission,attr"`
}

// Service represents a Service entry in the manifest file
type Service struct {
	Name       string `xml:"name,attr"`
	IsExported string `xml:"exported,attr"`
	Permission string `xml:"permission,attr"`
}

// ApplicationInfo holds all the data about the application components of the app
type ApplicationInfo struct {
	Name               string              `xml:"name,attr"`
	AllowADBBackup     string              `xml:"allowBackup,attr"`
	Activities         []Activity          `xml:"activity"`
	BroadcastReceivers []BroadcastReceiver `xml:"receiver"`
	Services           []Service           `xml:"service"`
}

// Manifest is a marshaled version of all the data in the AndroidManifest.xml file
type Manifest struct {
	PackageName string          `xml:"package,attr"`
	SDKInfo     SDKInfo         `xml:"uses-sdk"`
	Application ApplicationInfo `xml:"application"`
	Permissions []Permission    `xml:"uses-permission"`
}

type ManifestUnit struct {
	Document *xmlquery.Node
}

func (unit ManifestUnit) Type() engine.UnitType {
	return engine.StructuredDataUnit
}

// nolint Complex method for pass refactor now TODO: Refactor this method in the future to clean code
func (unit ManifestUnit) Eval(rule engine.Rule) (unitFindings []engine.Finding) {
	if structuredDataRule, ok := rule.(platforms.StructuredDataRule); ok {
		switch structuredDataRule.Type {
		case platforms.RegularMatch:
			for _, expression := range structuredDataRule.Expressions {
				exprResult := xmlquery.QuerySelectorAll(unit.Document, expression)

				if len(exprResult) < 1 {
					return
				}

				for _, finding := range exprResult {
					unitFindings = append(
						unitFindings,
						platforms.PopulateFindingWithRuleMetadata(
							structuredDataRule,
							"AndroidManifest.xml", finding.OutputXML(true), 0, 0),
					)
				}
			}
		case platforms.NotMatch:
			for _, expression := range structuredDataRule.Expressions {
				exprResult := xmlquery.QuerySelectorAll(unit.Document, expression)

				if len(exprResult) > 0 {
					return
				}

				unitFindings = append(
					unitFindings,
					platforms.PopulateFindingWithRuleMetadata(structuredDataRule,
						"AndroidManifest.xml", "", 0, 0),
				)
			}
		default:
			return unitFindings
		}
	}

	return unitFindings
}

func NewManifestUnit(content []byte) (unit *ManifestUnit, err error) {
	manifestRawDataReader := bytes.NewReader(content)

	formattedDocument, err := xmlquery.Parse(manifestRawDataReader)

	if err != nil {
		return unit, err
	}

	unit = &ManifestUnit{
		Document: formattedDocument,
	}

	return unit, nil
}
