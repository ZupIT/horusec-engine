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

package cst

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/ZupIT/horusec-devkit/pkg/enums/languages"
	treesitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
)

// Visitor A Visitor's Visit method is invoked for each node encountered by Walk.
// If the result visitor w is not nil, Walk visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).
type Visitor interface {
	// Visit visit node in CST
	Visit(*Node) Visitor
}

// Walk traverses an CST in depth-first order: It starts by calling
// v.Visit(node); node must not be nil. If the visitor w returned by
// v.Visit(node) is not nil, Walk is invoked recursively with visitor
// w for each of the non-nil children of node, followed by a call of
// w.Visit(nil).
// v.Accept will be called for each child of node, if v.Accept returns
// false, the iteration will skip the next child node of current node
//
// Example:
// Consider the following CST:
//
// (statement_block (comment) (lexical_declaration (variable_declarator name: (identifier) value: (string)))
// (lexical_declaration (variable_declarator name: (identifier) value: (string))))
//
// If v.Accept returns false for comment node, the following lexical_declaration will be skipped and the second
// lexical_declaration will be passed to v.Visit
func Walk(v Visitor, node *Node) {
	if v = v.Visit(node); v == nil {
		return
	}
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		Walk(v, child)
	}

	v.Visit(nil)
}

type inspector func(*Node) bool

func (f inspector) Visit(node *Node) Visitor {
	if f(node) {
		return f
	}

	return nil
}

// Inspect traverses an CST in depth-first order: It starts by calling
// f(node); node must not be nil. If f returns true, Inspect invokes f
// recursively for each of the non-nil children of node, followed by a
// call of f(nil).
//
func Inspect(node *Node, f func(*Node) bool) {
	Walk(inspector(f), node)
}

// Parse parse a src into a tree and return the root node of the tree.
// The src should be a valid code
//
// nolint:funlen,exhaustive // We don't support all languages yet.
func Parse(src []byte, language languages.Language) (*Node, error) {
	parser := treesitter.NewParser()

	switch language {
	case languages.Javascript:
		parser.SetLanguage(javascript.GetLanguage())
	case languages.Java:
		parser.SetLanguage(java.GetLanguage())
	default:
		return nil, errors.New("invalid language")
	}

	node, err := parser.ParseCtx(context.Background(), nil, src)
	if err != nil {
		return nil, fmt.Errorf("parse node: %w", err)
	}

	return newNode(node.RootNode(), src), nil
}

// Node is a wrapper around treesitter.Node that holds the source
// code used to create the initial CST.
//
// This source code is used to get the source code reprensentation
// of some node inside CST.
type Node struct {
	node *treesitter.Node
	src  []byte
}

func newNode(node *treesitter.Node, src []byte) *Node {
	return &Node{
		node: node,
		src:  src,
	}
}

// NamedChild returns the node's *named* child at the given index.
func (n *Node) NamedChild(idx int) *Node {
	if child := n.node.NamedChild(idx); child != nil {
		return newNode(child, n.src)
	}

	return nil
}

// NamedChildCount returns the node's number of *named* children.
func (n *Node) NamedChildCount() int {
	return int(n.node.NamedChildCount())
}

// ChildByFieldName returns the node's child with the given field name.
func (n *Node) ChildByFieldName(name string) *Node {
	child := n.node.ChildByFieldName(name)
	if child != nil {
		return newNode(child, n.src)
	}

	return nil
}

// Parent returns the node's immediate Parent.
func (n *Node) Parent() *Node {
	if p := n.node.Parent(); p != nil {
		return newNode(p, n.src)
	}

	return nil
}

// Value return the bytes value of node
func (n *Node) Value() []byte {
	return n.src[n.node.StartByte():n.node.EndByte()]
}

// Type returns the node type as a string.
func (n *Node) Type() string {
	return n.node.Type()
}

// IsError Check if this node represents a syntax error.
//
// Syntax errors represent parts of the code that could not be incorporated into a
// valid syntax tree.
func (n *Node) IsError() bool {
	// This code is copied from Rust bindings implementation
	// https://docs.rs/tree-sitter/0.19.5/tree_sitter/struct.Node.html#method.is_error
	return n.node.Symbol() == math.MaxUint16
}

// StartByte returns the node's start byte.
func (n *Node) StartByte() uint32 {
	return n.node.StartByte()
}

// EndByte returns the node's end byte.
func (n *Node) EndByte() uint32 {
	return n.node.EndByte()
}

// StartPoint returns the node's start position in terms of rows and columns.
func (n *Node) StartPoint() treesitter.Point {
	return n.node.StartPoint()
}

// EndPoint returns the node's end position in terms of rows and columns.
func (n *Node) EndPoint() treesitter.Point {
	return n.node.EndPoint()
}

// String returns an S-expression representing the node as a string.
func (n *Node) String() string {
	return n.node.String()
}

// SanitizeNodeValue removes ' and " prefix/suffix for a byte slice
func SanitizeNodeValue(b []byte) []byte {
	if bytes.HasPrefix(b, []byte("'")) || bytes.HasPrefix(b, []byte(`"`)) || bytes.HasPrefix(b, []byte("`")) {
		return b[1 : len(b)-1]
	}

	return b
}

// IterNamedChilds iterate over named childs from node
// calling fn using each named child node from iteration.
func IterNamedChilds(node *Node, fn func(node *Node)) {
	for idx := 0; idx < node.NamedChildCount(); idx++ {
		fn(node.NamedChild(idx))
	}
}
