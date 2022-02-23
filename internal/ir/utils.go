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

package ir

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/ZupIT/horusec-engine/internal/ast"
)

// debug is a variable defined at compile time that informs if
// this package was compiled in debug mode. Default is false.
// Accepted values are 1 or 0.
//
// debug mode enabled means that panic's are expected.
var debug = "0"

// unsupportedNode is a function that handle unsupported nodes.
//
// This function should be called on *every* place when we are not
// handling all scenarios of AST to IR conversions.
//
// unsupportedNode has two implementations; one for development that
// will panic on any ast.Node not supported and other for production
// environments that will only skip the node.
//
// The main goal of this function is to provide a way to in development
// catch AST nodes that is not supported yet and also be possible to
// integrate the experimental semantic analysis on horusec cli without
// stoping the analysis with a panic.
//
// The function implementation is configured at compile time using the
// flag -ldflags "-X github.com/ZupIT/horusec-engine/internal/ir.debug=1"
//
// The default value is 0, which means that debug is disabled.
//
// Clients should *always* call unsupportedNode and not their implementation.
var unsupportedNode func(ast.Node)

// nolint: gochecknoinits // init is necessary to set the unsupportedNode handler.
func init() {
	if debug == "1" {
		unsupportedNode = _panicUnsupportedNode
	} else {
		unsupportedNode = _skipUnsupportedNode
	}
}

// debugIsEnable return true if ir package was compiled in debug mode.
func debugIsEnable() bool {
	return debug == "1"
}

// _skipUnsupportedNode is a implementation of unsupportedNode var that just skip a
// not supported ast.Node.
//
// NOTE: Never call this function directly, you should call unsupportedNode instead.
func _skipUnsupportedNode(node ast.Node) {}

// panicUnsupportedNode is a implementation of unsupportedNode var that panic
// for unsupported nodes.
//
// The panic message will contains the caller of panicUnsupportedNode,
// the starting and ending position of the node.
//
// NOTE: Never call this function directly, you should call unsupportedNode instead.
func _panicUnsupportedNode(node ast.Node) {
	buf := bytes.NewBufferString("")

	if pc, _, _, ok := runtime.Caller(1); ok {
		if caller := runtime.FuncForPC(pc); caller != nil {
			fmt.Fprintf(buf, "%s: ", caller.Name())
		}
	}

	fmt.Fprintf(buf, "unsupported node %T", node)

	pos := node.Pos()
	panic(fmt.Sprintf("%s\nStart:\t<%s>\nEnd:\t<%s>", buf.String(), pos.Start(), pos.End()))
}
