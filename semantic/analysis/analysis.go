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

package analysis

import (
	"github.com/ZupIT/horusec-engine/internal/ast"
	"github.com/ZupIT/horusec-engine/internal/ir"
)

// Analyzer represents an analysis entrypoint that will analyze a single
// file on IR representation.
type Analyzer interface {
	Run(*Pass)
}

// AnalyzerValue represents an analysis check that analyze ir Values.
type AnalyzerValue interface {
	Run(ir.Value) bool
}

// Pass describes a single unit of work: the application of a particular Analyzer
// to a particular IR file. The Pass provides information to the Analyzer's interface
// about the file being analyzed, and provides operations to the Analyzer for reporting
// issues.
type Pass struct {
	File     *ir.File     // Current file being analyzed.
	Function *ir.Function // Current function being analyzed.

	// Report reports a Issue, a vulnerability about a specific
	// location in the analyzed source code.
	Report func(Issue)
}

// Issue represents a vulnerability issue on a source code.
//
// The source code vulnerability could be accessed using StartOffset
// and EndOffset on the raw bytes file content.
type Issue struct {
	Filename    string // Name of vulenrable file.
	StartOffset int    // Start byte offset of source code.
	EndOffset   int    // End byte offset of source code.
	Line        int    // Start line of source code.
	Column      int    // Start column of source code.
}

// NewIssue create a new issue to a given node.
func NewIssue(filename string, node ast.Node) Issue {
	pos := node.Pos()
	start, end := pos.Start(), pos.End()

	return Issue{
		Filename:    filename,
		StartOffset: int(start.Byte),
		EndOffset:   int(end.Byte),
		Line:        int(start.Row),
		Column:      int(start.Column),
	}
}
