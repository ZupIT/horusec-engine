// Copyright 2020 ZUP IT SERVICOS EM TECNOLOGIA E INOVACAO SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package engine

// UnitType defines which type of content, and therefore, which kind of rule
// is needed in order to extract information about the program we are analyzing
type UnitType int

const (
	ProgramTextUnit UnitType = iota
	StructuredDataUnit
)

// Rule defines a generic rule for any kind of analysis the engine have to execute
type Rule interface {
	IsFor(UnitType) bool // Indicates which kind of program unit this rules can be ran on
}

// Metadata holds information for the rule to match a useful advisory
type Metadata struct {
	ID          string
	Name        string
	CodeSample  string
	Description string

	// Metadata levels
	Severity   string
	Confidence string
}
