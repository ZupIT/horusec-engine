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
	p := parser{}

	return p.parseCST(name, root), nil
}

type parser struct {
}

// parseCST parse a tree-sitter CST to a generic AST.
func (p *parser) parseCST(name string, root *cst.Node) *ast.File {
	file := &ast.File{
		Name: &ast.Ident{
			Name: name,
		},
	}

	// Here we just traverse the top level child nodes and let
	// the underlying parsing methods travese the sub nodes.
	cst.IterNamedChilds(root, func(node *cst.Node) {
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
		case ExpressionStatement:
			file.Exprs = append(file.Exprs, p.parseExpr(node))
		default:
			panic(fmt.Sprintf("unexpected node type %s", node.Type()))
		}
	})

	return file
}

func (p *parser) parseClassDecl(node *cst.Node) ast.Decl {
	classDecl := &ast.ClassDecl{
		Name: ast.NewIdent(node.ChildByFieldName("name")),
	}

	if body := node.ChildByFieldName("body"); body != nil {
		classDecl.Body = p.parseClassBody(body)
	}

	return classDecl
}

func (p *parser) parseClassBody(body *cst.Node) *ast.BodyDecl {
	var bodyDecl ast.BodyDecl

	cst.IterNamedChilds(body, func(node *cst.Node) {
		switch node.Type() {
		case PublicFieldDefinition:
			valueDecl := ast.ValueDecl{
				Names: []*ast.Ident{ast.NewIdent(node.ChildByFieldName("property"))},
			}
			if value := node.ChildByFieldName("value"); value != nil {
				valueDecl.Values = append(valueDecl.Values, p.parseExpr(value))
			}
			bodyDecl.List = append(bodyDecl.List, &valueDecl)
		case MethodDefinition:
			bodyDecl.List = append(bodyDecl.List, p.parseFuncDecl(node))
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
		Name: ident,
		Type: new(ast.FuncType),
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

	cst.IterNamedChilds(node, func(node *cst.Node) {
		stmts = append(stmts, p.parseStmt(node))
	})

	return &ast.BlockStmt{
		List: stmts,
	}
}

func (p *parser) parseParameters(node *cst.Node) *ast.FieldList {
	var parameters []*ast.Field

	cst.IterNamedChilds(node, func(node *cst.Node) {
		parameters = append(parameters, &ast.Field{
			Name: p.parseExpr(node),
		})
	})
	return &ast.FieldList{
		List: parameters,
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
	var varDecl ast.ValueDecl

	cst.IterNamedChilds(node, func(node *cst.Node) {
		assertNodeType(node, VariableDeclarator)

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
				}
				return
			default:
				// Otherwise we just parse value as an expression.
				varDecl.Values = append(varDecl.Values, p.parseExpr(value))
			}
		}

		assertNodeType(name, Identifier)
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
	case LexicalDeclaration:
		lhs := make([]ast.Expr, 0, node.NamedChildCount())
		rhs := make([]ast.Expr, 0, node.NamedChildCount())

		cst.IterNamedChilds(node, func(node *cst.Node) {
			name := node.ChildByFieldName("name")
			assertNodeType(name, Identifier)
			lhs = append(lhs, ast.NewIdent(name))

			if value := node.ChildByFieldName("value"); value != nil {
				rhs = append(rhs, p.parseExpr(value))
			}
		})

		return &ast.AssignStmt{
			LHS: lhs,
			RHS: rhs,
		}
	case ExpressionStatement:
		// We need to check the child of expression_statement and check
		// their type to try to convert to some non-generic type of statement.
		switch child := node.NamedChild(0); child.Type() {
		case AssignmentExpression:
			left := child.ChildByFieldName("left")
			right := child.ChildByFieldName("right")
			return &ast.AssignStmt{
				LHS: []ast.Expr{p.parseExpr(left)},
				RHS: []ast.Expr{p.parseExpr(right)},
			}
		default:
			return &ast.ExprStmt{
				Expr: p.parseExpr(child),
			}
		}
	case ReturnStatement:
		expr := node.NamedChild(0)
		if expr.Type() == SequenceExpression {
			return &ast.ReturnStmt{
				Results: p.parseSequenceExpr(expr),
			}
		}
		return &ast.ReturnStmt{
			Results: []ast.Expr{
				p.parseExpr(expr),
			},
		}
	case IfStatement:
		stmt := &ast.IfStmt{
			Cond: p.parseExpr(node.ChildByFieldName("condition")),
			Body: p.parseFuncBody(node.ChildByFieldName("consequence")),
		}

		if alt := node.ChildByFieldName("alternative"); alt != nil {
			stmt.Else = p.parseStmt(alt)
		}
		return stmt
	case ElseClause, StatementBlock:
		// Here we just need to get the first named child that is an statement.
		if child := node.NamedChild(0); child != nil {
			// We can have an else branch without body.
			return p.parseStmt(child)
		}
		return nil
	default:
		panic(fmt.Sprintf("not handled statement of type <%s>", node.Type()))
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
			Kind:  node.Type(),
			Value: string(cst.SanitizeNodeValue(node.Value())),
		}
	case True, False:
		return &ast.BasicLit{
			Kind:  "boolean",
			Value: string(node.Value()),
		}
	case NewExpression:
		parent := node.Parent()
		assertNodeType(parent, VariableDeclarator)

		var args []ast.Expr
		cst.IterNamedChilds(node.ChildByFieldName("arguments"), func(node *cst.Node) {
			args = append(args, p.parseExpr(node))
		})

		name := parent.ChildByFieldName("name")
		assertNodeType(name, Identifier)

		return &ast.ObjectExpr{
			Name: ast.NewIdent(name),
			Type: p.parseExpr(node.ChildByFieldName("constructor")),
			Elts: args,
		}
	case Object:
		var obj ast.ObjectExpr
		cst.IterNamedChilds(node, func(pair *cst.Node) {
			obj.Elts = append(obj.Elts, p.parseExpr(pair))
		})
		return &obj
	case Pair:
		return &ast.KeyValueExpr{
			Key:   p.parseExpr(node.ChildByFieldName("key")),
			Value: p.parseExpr(node.ChildByFieldName("value")),
		}
	case Array:
		var obj ast.ObjectExpr
		cst.IterNamedChilds(node, func(node *cst.Node) {
			obj.Elts = append(obj.Elts, p.parseExpr(node))
		})
		return &obj
	case BinaryExpression:
		// TODO: tree-sitter doesn't expose a operator node
		// we need to find a way to get this operator.
		return &ast.BinaryExpr{
			Left:  p.parseExpr(node.ChildByFieldName("left")),
			Right: p.parseExpr(node.ChildByFieldName("right")),
		}
	case ParenthesizedExpression:
		return p.parseExpr(node.NamedChild(0))
	case AssignmentPattern:
		left := node.ChildByFieldName("left")
		assertNodeType(left, Identifier)

		return &ast.ObjectExpr{
			Name: ast.NewIdent(left),
			Type: nil,
			Elts: []ast.Expr{p.parseExpr(node.ChildByFieldName("right"))},
		}
	case CallExpression:
		fn := node.ChildByFieldName("function")
		args := node.ChildByFieldName("arguments")

		argsExpr := make([]ast.Expr, 0, args.NamedChildCount())
		cst.IterNamedChilds(args, func(node *cst.Node) {
			argsExpr = append(argsExpr, p.parseExpr(node))
		})

		return &ast.CallExpr{
			Fun:  p.parseExpr(fn),
			Args: argsExpr,
		}
	case MemberExpression:
		return &ast.SelectorExpr{
			Expr: p.parseExpr(node.ChildByFieldName("object")),
			Sel:  ast.NewIdent(node.ChildByFieldName("property")),
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
			Body: p.parseFuncBody(node.ChildByFieldName("body")),
		}
	case TemplateString:
		var exprs []ast.Expr
		cst.IterNamedChilds(node, func(node *cst.Node) {
			exprs = append(exprs, p.parseExpr(node.NamedChild(0)))
		})
		return &ast.TemplateExpr{
			Value: string(cst.SanitizeNodeValue(node.Value())),
			Subs:  exprs,
		}
	case ExpressionStatement:
		// ExpressionStatement is composed only by an expression and semicolon
		// so here we just need to get and parse the expression.
		return p.parseExpr(node.NamedChild(0))
	default:
		panic(fmt.Sprintf("not handled expression of type <%s>", node.Type()))
	}
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
		assertNodeType(decl, VariableDeclarator)
		if args := node.ChildByFieldName("arguments"); args != nil {
			if args.NamedChildCount() > 0 {
				return &ast.ImportDecl{
					Path: ast.NewIdent(args.NamedChild(0)),
					Name: ast.NewIdent(decl.ChildByFieldName("name")),
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
				Path: ast.NewIdent(source),
			},
		}
	}

	switch imported := importClause.NamedChild(0); imported.Type() {
	case Identifier:
		// Handle `import name from 'module-name'`
		return []ast.Decl{
			&ast.ImportDecl{
				Name: ast.NewIdent(imported),
				Path: ast.NewIdent(source),
			},
		}
	case NamedImports:
		// Handle `import { name } from 'module-name'`
		// and `import { name as alias } from 'module-name'`
		var imports []ast.Decl
		cst.IterNamedChilds(imported, func(importIdentifier *cst.Node) {
			if name := importIdentifier.ChildByFieldName("name"); name != nil {
				imports = append(imports, &ast.ImportDecl{
					Name: ast.NewIdent(name),
					Path: ast.NewIdent(source),
				})
			}
		})
		return imports
	case NamespaceImport:
		// Handle `import * as name from 'module-name'`
		return []ast.Decl{
			&ast.ImportDecl{
				Name: ast.NewIdent(imported.NamedChild(0)),
				Path: ast.NewIdent(source),
			},
		}
	}
	return []ast.Decl{}
}

func assertNodeType(node *cst.Node, typ string) {
	if node.Type() != typ {
		panic(fmt.Sprintf("Expected <%s> node, got <%s>", typ, node.Type()))
	}
}
