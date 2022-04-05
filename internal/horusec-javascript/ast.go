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

// nolint:funlen,gocyclo,gocognit // We need a lot of lines and if's to parse CST to AST
package javascript

import (
	"bytes"
	"fmt"

	"github.com/ZupIT/horusec-devkit/pkg/enums/languages"

	"github.com/ZupIT/horusec-engine/internal/ast"
	"github.com/ZupIT/horusec-engine/internal/cst"
)

// ParseFile parses the source code of a single JavaScript source file
// and returns the corresponding ast.File node.
func ParseFile(name string, src []byte) (*ast.File, error) {
	root, err := cst.Parse(src, languages.Javascript)
	if err != nil {
		return nil, err
	}
	p := parser{
		name: name,
	}

	return p.parseCST(name, root), nil
}

type parser struct {
	name string // Name of file being parsed.
}

// parseCST parse a tree-sitter CST to a generic AST.
func (p *parser) parseCST(name string, root *cst.Node) *ast.File {
	file := &ast.File{
		Position: ast.NewPosition(root),
		Name: &ast.Ident{
			Name: name,
		},
	}

	if p.isNosecFile(root) {
		return file
	}

	ignoreNextNode := false

	// Here we just traverse the top level child nodes and let
	// the underlying parsing methods travese the sub nodes.
	cst.IterNamedChilds(root, func(node *cst.Node) {
		if ignoreNextNode {
			ignoreNextNode = false
			return
		}

		switch node.Type() {
		// Top-level declarations
		case VariableDeclaration, LexicalDeclaration:
			file.Decls = append(file.Decls, p.parseVarDecl(node)...)
		case FunctionDeclaration:
			file.Decls = append(file.Decls, p.parseFuncDecl(node))
		case ImportStatement:
			file.Decls = append(file.Decls, p.parseImportStmt(node)...)
		case ClassDeclaration:
			file.Decls = append(file.Decls, p.parseClassDecl(node))

		// Top-level expressions
		// TODO: check for more top level statements
		case ExpressionStatement:
			file.Exprs = append(file.Exprs, p.parseExpr(node))

		case Comment:
			ignoreNextNode = ast.IsNosec(node.Value())
		default:
			file.BadNodes = append(file.BadNodes, ast.NewUnsupportedNode(node))
		}
	})

	return file
}

// isNosecFile return true is the entire source file should be ignored.
func (p *parser) isNosecFile(root *cst.Node) bool {
	return root.NamedChildCount() > 0 && p.isNosec(root.NamedChild(0))
}

// isNosec return true is the given node is a comment with nosec directive.
func (p *parser) isNosec(node *cst.Node) bool {
	return node.Type() == Comment && ast.IsNosec(node.Value())
}

func (p *parser) parseClassDecl(node *cst.Node) ast.Decl {
	classDecl := &ast.ClassDecl{
		Name:     ast.NewIdent(node.ChildByFieldName("name")),
		Position: ast.NewPosition(node),
	}

	if body := node.ChildByFieldName("body"); body != nil {
		classDecl.Body = p.parseClassBody(body)
	}

	return classDecl
}

func (p *parser) parseClassBody(body *cst.Node) *ast.BodyDecl {
	bodyDecl := ast.BodyDecl{
		Position: ast.NewPosition(body),
	}

	ignoreNextNode := false
	cst.IterNamedChilds(body, func(node *cst.Node) {
		if ignoreNextNode {
			ignoreNextNode = false
			return
		}

		switch node.Type() {
		case PublicFieldDefinition:
			valueDecl := ast.ValueDecl{
				Names:    []*ast.Ident{ast.NewIdent(node.ChildByFieldName("property"))},
				Position: ast.NewPosition(node),
			}
			if value := node.ChildByFieldName("value"); value != nil {
				valueDecl.Values = append(valueDecl.Values, p.parseExpr(value))
			}
			bodyDecl.List = append(bodyDecl.List, &valueDecl)
		case MethodDefinition:
			bodyDecl.List = append(bodyDecl.List, p.parseFuncDecl(node))
		case Comment:
			ignoreNextNode = ast.IsNosec(node.Value())
		default:
			panic(fmt.Sprintf("unexpected class_definition child node: %s", node.Type()))
		}
	})

	return &bodyDecl
}

func (p *parser) parseFuncDecl(node *cst.Node) ast.Decl {
	return p.parseArrowFunc(ast.NewIdent(node.ChildByFieldName("name")), node)
}

func (p *parser) parseArrowFunc(ident *ast.Ident, node *cst.Node) ast.Decl {
	decl := &ast.FuncDecl{
		Name:     ident,
		Type:     new(ast.FuncType),
		Position: ast.NewPosition(node),
	}

	if parameters := node.ChildByFieldName("parameters"); parameters != nil {
		decl.Type.Params = p.parseParameters(parameters)
	}

	if body := node.ChildByFieldName("body"); body != nil {
		decl.Body = p.parseFuncBody(body)
	}

	return decl
}

func (p *parser) parseFuncBody(node *cst.Node) *ast.BlockStmt {
	stmts := make([]ast.Stmt, 0, node.NamedChildCount())

	ignoreNextNode := false
	cst.IterNamedChilds(node, func(node *cst.Node) {
		if ignoreNextNode {
			ignoreNextNode = false
			return
		}

		if p.isNosec(node) {
			ignoreNextNode = true
			return
		}

		stmts = append(stmts, p.parseStmt(node))
	})

	return &ast.BlockStmt{
		List: stmts,
	}
}

func (p *parser) parseParameters(node *cst.Node) *ast.FieldList {
	var parameters []*ast.Field

	p.iterNamedChilds(node, func(node *cst.Node) {
		parameters = append(parameters, &ast.Field{
			Name:     p.parseExpr(node),
			Position: ast.NewPosition(node),
		})
	})

	return &ast.FieldList{
		List:     parameters,
		Position: ast.NewPosition(node),
	}
}

func (p *parser) parseVarDecl(node *cst.Node) []ast.Decl {
	// variable_declaration node can have multiple variable_declarator nodes, so
	// we need to iterate over all of them because we can have multiple variable
	// declarations on the same node.
	//
	// Look for variable_declaration rule on:
	// https://github.com/tree-sitter/tree-sitter-javascript/blob/master/grammar.js
	//
	// TODO: Here we are generating a unique ast.ValueDecl for multiple variable_declaration
	// we need to convert multi variable_declaration in multi ast.ValueDecl values.
	// Example:
	// const a = 1, b = 2; Should generate two instances of ast.ValueDecl, one for each
	// variable_declaration.
	// NOTE: We **should** not convert tuple declaration to multi ast.ValueDecl
	// Example:
	// const a, b = foo(); Should generate one ast.ValueDecl with two Names and one Value.

	var decls []ast.Decl
	varDecl := ast.ValueDecl{
		Position: ast.NewPosition(node),
	}

	p.iterNamedChilds(node, func(node *cst.Node) {
		p.assertNodeType(node, VariableDeclarator)

		name := node.ChildByFieldName("name")

		if value := node.ChildByFieldName("value"); value != nil {
			switch value.Type() {
			case ArrowFunction:
				// If value of variable_declaration is an arrow_function
				// we need to convert to a normal function.
				decls = append(decls, p.parseArrowFunc(ast.NewIdent(name), value))

				return
			case CallExpression:
				// If value of variable_declaration is a function call to
				// require, we need to convert to a import_statement.
				if decl := p.parseRequireCallExpr(value); decl != nil {
					decls = append(decls, decl)

					return
				}

				// Otherwise just parse as a normal call expression.
				varDecl.Values = append(varDecl.Values, p.parseCallExpr(value))
			default:
				// Otherwise we just parse value as an expression.
				varDecl.Values = append(varDecl.Values, p.parseExpr(value))
			}
		}

		p.assertNodeType(name, Identifier)
		varDecl.Names = append(varDecl.Names, ast.NewIdent(name))
	})

	// Just add variable declaration if we had one.
	if len(varDecl.Names) > 0 {
		decls = append(decls, &varDecl)
	}

	return decls
}

func (p *parser) parseStmt(node *cst.Node) ast.Stmt {
	switch node.Type() {
	case LexicalDeclaration, VariableDeclaration:
		lhs := make([]ast.Expr, 0, node.NamedChildCount())
		rhs := make([]ast.Expr, 0, node.NamedChildCount())

		p.iterNamedChilds(node, func(node *cst.Node) {
			name := node.ChildByFieldName("name")
			p.assertNodeType(name, Identifier)
			lhs = append(lhs, ast.NewIdent(name))

			if value := node.ChildByFieldName("value"); value != nil {
				rhs = append(rhs, p.parseExpr(value))
			}
		})

		return &ast.AssignStmt{
			LHS:      lhs,
			RHS:      rhs,
			Position: ast.NewPosition(node),
		}
	case ExpressionStatement:
		// We need to check the child of expression_statement and check
		// their type to try to convert to some non-generic type of statement.
		switch child := node.NamedChild(0); child.Type() {
		case AssignmentExpression:
			left := child.ChildByFieldName("left")
			right := child.ChildByFieldName("right")

			return &ast.AssignStmt{
				LHS:      []ast.Expr{p.parseExpr(left)},
				RHS:      []ast.Expr{p.parseExpr(right)},
				Position: ast.NewPosition(node),
			}
		default:
			return &ast.ExprStmt{
				Expr:     p.parseExpr(child),
				Position: ast.NewPosition(node),
			}
		}
	case ReturnStatement:
		if expr := node.NamedChild(0); expr != nil {
			if expr.Type() == SequenceExpression {
				return &ast.ReturnStmt{
					Results:  p.parseSequenceExpr(expr),
					Position: ast.NewPosition(node),
				}
			}

			return &ast.ReturnStmt{
				Position: ast.NewPosition(expr),
				Results: []ast.Expr{
					p.parseExpr(expr),
				},
			}
		}

		// Return without any statement
		return &ast.ReturnStmt{
			Position: ast.NewPosition(node),
			Results:  make([]ast.Expr, 0),
		}
	case IfStatement:
		stmt := &ast.IfStmt{
			Cond:     p.parseExpr(node.ChildByFieldName("condition")),
			Body:     p.parseFuncBody(node.ChildByFieldName("consequence")),
			Position: ast.NewPosition(node),
		}

		if alt := node.ChildByFieldName("alternative"); alt != nil {
			stmt.Else = p.parseStmt(alt)
		}

		return stmt
	case ElseClause:
		// Here we just need to get the first named child that is a statement.
		if child := node.NamedChild(0); child != nil {
			// We can have an else branch without body.
			return p.parseStmt(child)
		}

		return nil
	case StatementBlock:
		block := &ast.BlockStmt{
			Position: ast.NewPosition(node),
			List:     make([]ast.Stmt, 0, node.NamedChildCount()),
		}
		p.iterNamedChilds(node, func(node *cst.Node) {
			block.List = append(block.List, p.parseStmt(node))
		})

		return block
	case TryStatement:
		stmt := &ast.TryStmt{
			Position: ast.NewPosition(node),
			Body:     p.parseFuncBody(node.ChildByFieldName("body")),
		}

		if catchClause := node.ChildByFieldName("handler"); catchClause != nil {
			stmt.CatchClause = []*ast.CatchClause{
				{
					Position:  ast.NewPosition(catchClause),
					Parameter: ast.NewIdent(catchClause.ChildByFieldName("parameter")),
					Body:      p.parseFuncBody(catchClause.ChildByFieldName("body")),
				},
			}
		}

		if finalizer := node.ChildByFieldName("finalizer"); finalizer != nil {
			stmt.Finalizer = p.parseFuncBody(finalizer.ChildByFieldName("body"))
		}

		return stmt
	case WhileStatement:
		stmt := &ast.WhileStmt{
			Position: ast.NewPosition(node),
			Cond:     p.parseExpr(node.ChildByFieldName("condition")),
			Body:     p.parseFuncBody(node.ChildByFieldName("body")),
		}

		return stmt
	case SwitchStatement:
		stmt := &ast.SwitchStatement{
			Position: ast.NewPosition(node),
			Value:    p.parseExpr(node.ChildByFieldName("value")),
			Body:     p.parseFuncBody(node.ChildByFieldName("body")),
		}

		return stmt
	case SwitchCase:
		stmt := &ast.SwitchCase{
			Position: ast.NewPosition(node),
			Cond:     p.parseExpr(node.ChildByFieldName("value")),
		}

		body := make([]ast.Stmt, 0, node.NamedChildCount()-1)

		// The first named child of switch case is the condition, here we just need to iterate over the
		// statements of switch case. The condition of switch case is parsed above.
		for i := 1; i < node.NamedChildCount(); i++ {
			body = append(body, p.parseStmt(node.NamedChild(i)))
		}

		stmt.Body = body

		return stmt
	case SwitchDefault:
		stmt := &ast.SwitchDefault{
			Position: ast.NewPosition(node),
		}

		body := make([]ast.Stmt, 0, node.NamedChildCount())

		p.iterNamedChilds(node, func(node *cst.Node) {
			body = append(body, p.parseStmt(node))
		})

		stmt.Body = body

		return stmt
	case BreakStatement:
		stmt := &ast.BreakStatement{
			Position: ast.NewPosition(node),
		}

		if label := node.ChildByFieldName("label"); label != nil {
			stmt.Label = ast.NewIdent(label)
		}

		return stmt
	case ForStatement:
		var incExpr ast.Expr
		if inc := node.ChildByFieldName("increment"); inc != nil {
			incExpr = p.parseExpr(inc)
		}

		return &ast.ForStatement{
			Position:  ast.NewPosition(node),
			VarDecl:   p.parseStmt(node.ChildByFieldName("initializer")),
			Cond:      p.parseExpr(node.ChildByFieldName("condition")),
			Increment: incExpr,
			Body:      p.parseFuncBody(node.ChildByFieldName("body")),
		}
	case ForInStatement:
		return &ast.ForInStatement{
			Position: ast.NewPosition(node),
			Left:     p.parseExpr(node.ChildByFieldName("left")),
			Right:    p.parseExpr(node.ChildByFieldName("right")),
			Body:     p.parseFuncBody(node.ChildByFieldName("body")),
		}

	case ContinueStatement:
		stmt := &ast.ContinueStatement{
			Position: ast.NewPosition(node),
		}

		if label := node.ChildByFieldName("label"); label != nil {
			stmt.Label = ast.NewIdent(label)
		}

		return stmt
	case LabeledStatement:
		stmt := &ast.LabeledStatement{
			Position: ast.NewPosition(node),
		}
		body := make([]ast.Stmt, 0, node.NamedChildCount())

		// The first named child of switch case is the condition, here we just need to iterate over the
		// statements of switch case. The condition of switch case is parsed above.
		for i := 1; i < node.NamedChildCount(); i++ {
			body = append(body, p.parseStmt(node.NamedChild(i)))
		}
		if label := node.ChildByFieldName("label"); label != nil {
			stmt.Label = ast.NewIdent(label)
		}
		stmt.Body = body

		return stmt
	case ExportStatement, EmptyStatement:
		// Since export statements will not be very useful information in our ast for now,
		// we will ignore this statement.
		return nil
	default:
		return ast.NewUnsupportedNode(node)
	}
}

func (p *parser) parseSequenceExpr(node *cst.Node) []ast.Expr {
	var exprs []ast.Expr

	left := node.ChildByFieldName("left")
	right := node.ChildByFieldName("right")

	// Left and Right values from sequence_expression can be
	// another sequence_expression, so we need to parse recursively.
	if left.Type() == SequenceExpression {
		exprs = append(exprs, p.parseSequenceExpr(left)...)
	} else {
		exprs = append(exprs, p.parseExpr(left))
	}

	if right.Type() == SequenceExpression {
		exprs = append(exprs, p.parseSequenceExpr(right)...)
	} else {
		exprs = append(exprs, p.parseExpr(right))
	}

	return exprs
}

func (p *parser) parseExpr(node *cst.Node) ast.Expr {
	switch node.Type() {
	case Identifier, PropertyIdentifier, This:
		return ast.NewIdent(node)
	case String, Number:
		return &ast.BasicLit{
			Kind:     node.Type(),
			Value:    string(cst.SanitizeNodeValue(node.Value())),
			Position: ast.NewPosition(node),
		}
	case True, False:
		return &ast.BasicLit{
			Kind:     "boolean",
			Value:    string(node.Value()),
			Position: ast.NewPosition(node),
		}
	case NewExpression:
		var name *ast.Ident
		if parent := node.Parent(); parent.Type() == VariableDeclarator {
			if n := parent.ChildByFieldName("name"); n != nil && n.Type() == Identifier {
				name = ast.NewIdent(n)
			}
		}

		var args []ast.Expr
		p.iterNamedChilds(node.ChildByFieldName("arguments"), func(node *cst.Node) {
			args = append(args, p.parseExpr(node))
		})

		return &ast.ObjectExpr{
			Name:     name,
			Type:     p.parseExpr(node.ChildByFieldName("constructor")),
			Elts:     args,
			Position: ast.NewPosition(node),
			Comment:  "constructor",
		}
	case Object:
		obj := ast.ObjectExpr{
			Comment: "hashmap",
		}

		p.iterNamedChilds(node, func(pair *cst.Node) {
			obj.Elts = append(obj.Elts, p.parseExpr(pair))
		})

		return &obj
	case Pair:
		return &ast.KeyValueExpr{
			Key:      p.keyFromPair(node.ChildByFieldName("key")),
			Value:    p.parseExpr(node.ChildByFieldName("value")),
			Position: ast.NewPosition(node),
		}
	case Array:
		obj := ast.ObjectExpr{
			Comment: "array",
		}

		p.iterNamedChilds(node, func(node *cst.Node) {
			obj.Elts = append(obj.Elts, p.parseExpr(node))
		})

		return &obj
	case BinaryExpression:
		return &ast.BinaryExpr{
			Left:     p.parseExpr(node.ChildByFieldName("left")),
			Right:    p.parseExpr(node.ChildByFieldName("right")),
			Op:       node.ChildByFieldName("operator").Type(), // Type will return the operador cleaned.
			Position: ast.NewPosition(node),
		}
	case ParenthesizedExpression:
		return p.parseExpr(node.NamedChild(0))
	case AssignmentPattern:
		left := node.ChildByFieldName("left")
		p.assertNodeType(left, Identifier)

		return &ast.ObjectExpr{
			Name:     ast.NewIdent(left),
			Type:     nil,
			Elts:     []ast.Expr{p.parseExpr(node.ChildByFieldName("right"))},
			Position: ast.NewPosition(node),
		}
	case CallExpression:
		return p.parseCallExpr(node)
	case MemberExpression:
		return &ast.SelectorExpr{
			Expr:     p.parseExpr(node.ChildByFieldName("object")),
			Sel:      ast.NewIdent(node.ChildByFieldName("property")),
			Position: ast.NewPosition(node),
		}
	case ArrowFunction, Function:
		var params *ast.FieldList

		if parameters := node.ChildByFieldName("parameters"); parameters != nil {
			params = p.parseParameters(parameters)
		}

		return &ast.FuncLit{
			Type: &ast.FuncType{
				Params:  params,
				Results: nil,
			},
			Body:     p.parseFuncBody(node.ChildByFieldName("body")),
			Position: ast.NewPosition(node),
		}
	case TemplateString:
		var exprs []ast.Expr
		p.iterNamedChilds(node, func(node *cst.Node) {
			exprs = append(exprs, p.parseExpr(node.NamedChild(0)))
		})

		return &ast.TemplateExpr{
			Value:    string(cst.SanitizeNodeValue(node.Value())),
			Subs:     exprs,
			Position: ast.NewPosition(node),
		}
	case ExpressionStatement:
		// ExpressionStatement is composed only by an expression and semicolon
		// so here we just need to get and parse the expression.
		return p.parseExpr(node.NamedChild(0))
	case UpdateExpression:
		return &ast.IncExpr{
			Position: ast.NewPosition(node),
			Op:       node.ChildByFieldName("operator").Type(), // Type will return the operador cleaned.
			Arg:      ast.NewIdent(node.ChildByFieldName("argument")),
		}
	case SubscriptExpression:
		return &ast.SubscriptExpr{
			Position: ast.NewPosition(node),
			Object:   p.parseExpr(node.ChildByFieldName("object")),
			Index:    p.parseExpr(node.ChildByFieldName("index")),
		}
	case EmptyStatement:
		return nil
	default:
		return ast.NewUnsupportedNode(node)
	}
}

// keyFromPair return the expression that represents a key from a pair node. If the key
// of pair is a property_indentifier keyFromPair return an ast.BasicLit with the value
// of property, this is necessary to avoid nil keys on IR, since the property identifier
// is commonly handled as a identifier.
func (p *parser) keyFromPair(node *cst.Node) ast.Expr {
	if node.Type() == PropertyIdentifier {
		return &ast.BasicLit{
			Position: ast.NewPosition(node),
			Kind:     "string",
			Value:    string(node.Value()),
		}
	}
	return p.parseExpr(node)
}

func (p *parser) parseRequireCallExpr(node *cst.Node) ast.Decl {
	// To extract the imports from require function, wee need to check
	// if a call expression is a call to require function, them we get
	// the argument from call and get the parent node that should be a
	// variable_declarator.
	// Then we get the first argument from require function call
	// and the name identifier from variable_declarator.
	//
	// TODO: We need to handle cases like: `const { foo, bar } = require('baz');`
	if fn := node.ChildByFieldName("function"); fn != nil && bytes.Equal(fn.Value(), []byte("require")) {
		decl := node.Parent()
		p.assertNodeType(decl, VariableDeclarator)
		if args := node.ChildByFieldName("arguments"); args != nil {
			if args.NamedChildCount() > 0 {
				return &ast.ImportDecl{
					Path:     ast.NewIdent(args.NamedChild(0)),
					Name:     ast.NewIdent(decl.ChildByFieldName("name")),
					Position: ast.NewPosition(node),
				}
			}
		}
	}

	return nil
}

func (p *parser) parseImportStmt(node *cst.Node) []ast.Decl {
	// On ImportStatetement we need get the ImportClause
	// that is the node that holds the modules that is being imported
	// If ImportClause is a string we just get the source import
	// of ImportStatement because the entire module is being imported
	// If is not a string, we need to get call Import that returns
	// the node that holds all modules that is being imported
	// If this node is an Identifier we just get your name
	// If is NamedImports we need iterate over all imports and
	// get your names and posible alias
	// If is NamespaceImport we just get the source import because
	// the entire module is being imported
	importClause := node.NamedChild(0)

	source := node.ChildByFieldName("source")

	// Handle `import './module-name'`
	if importClause.Type() == String {
		return []ast.Decl{
			&ast.ImportDecl{
				Path:     ast.NewIdent(source),
				Position: ast.NewPosition(node),
			},
		}
	}

	switch imported := importClause.NamedChild(0); imported.Type() {
	case Identifier:
		// Handle `import name from 'module-name'`
		return []ast.Decl{
			&ast.ImportDecl{
				Name:     ast.NewIdent(imported),
				Path:     ast.NewIdent(source),
				Position: ast.NewPosition(node),
			},
		}
	case NamedImports:
		// Handle `import { name } from 'module-name'`
		// and `import { name as alias } from 'module-name'`
		var imports []ast.Decl
		p.iterNamedChilds(imported, func(importIdentifier *cst.Node) {
			if name := importIdentifier.ChildByFieldName("name"); name != nil {
				imports = append(imports, &ast.ImportDecl{
					Name:     ast.NewIdent(name),
					Path:     ast.NewIdent(source),
					Alias:    p.aliasFromImportStmt(importIdentifier),
					Position: ast.NewPosition(node),
				})
			}
		})

		return imports
	case NamespaceImport:
		// Handle `import * as name from 'module-name'`
		return []ast.Decl{
			&ast.ImportDecl{
				Name:     ast.NewIdent(imported.NamedChild(0)),
				Path:     ast.NewIdent(source),
				Position: ast.NewPosition(node),
			},
		}
	}

	return []ast.Decl{}
}

// aliasFromImportStmt return an *ast.Ident with the alias name from import_specifier node.
//
// Return nil if import_specifier node don't have an alias.
func (p *parser) aliasFromImportStmt(node *cst.Node) *ast.Ident {
	if a := node.ChildByFieldName("alias"); a != nil {
		return ast.NewIdent(a)
	}

	return nil
}

func (p *parser) parseCallExpr(node *cst.Node) *ast.CallExpr {
	fn := node.ChildByFieldName("function")
	args := node.ChildByFieldName("arguments")

	argsExpr := make([]ast.Expr, 0, args.NamedChildCount())
	p.iterNamedChilds(args, func(node *cst.Node) {
		argsExpr = append(argsExpr, p.parseExpr(node))
	})

	return &ast.CallExpr{
		Fun:      p.parseExpr(fn),
		Args:     argsExpr,
		Position: ast.NewPosition(node),
	}
}

// iterNamedChilds iterate over named childs from node calling fn using each named
// child node from iteration ignoring nodes that are comments.
func (p *parser) iterNamedChilds(node *cst.Node, fn func(node *cst.Node)) {
	cst.IterNamedChildsIgnoringNode(node, Comment, fn)
}

func (p *parser) assertNodeType(node *cst.Node, typ string) {
	if node.Type() != typ {
		start := node.StartPoint()
		panic(fmt.Sprintf(
			"Expected <%s> node, got <%s> at %s:%d:%d", typ, node.Type(), p.name, start.Row, start.Column,
		))
	}
}
