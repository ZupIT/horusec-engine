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

package javascript_test

import (
	"testing"

	"github.com/ZupIT/horusec-engine/internal/ast"
	javascript "github.com/ZupIT/horusec-engine/internal/horusec-javascript"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testcase struct {
	name string
	src  string
	file ast.File
}

func TestParseClassDeclaration(t *testing.T) {
	testcases := []testcase{
		{
			name: "EmptyClassDeclaration",
			src:  `class Foo {}`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ClassDecl{
						Name: &ast.Ident{
							Name: "Foo",
						},
						Body: &ast.BodyDecl{},
					},
				},
			},
		},
		{
			name: "ClassDeclarationWithBody",
			src: `
class Foo{
  a;
  b = 10;
  constructor(a, b) {
    this.a = a;
    this.b = b;
  }
}
			`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ClassDecl{
						Name: &ast.Ident{
							Name: "Foo",
						},
						Body: &ast.BodyDecl{
							List: []ast.Decl{
								&ast.ValueDecl{
									Names: []*ast.Ident{
										{
											Name: "a",
										},
									},
								},
								&ast.ValueDecl{
									Names: []*ast.Ident{
										{
											Name: "b",
										},
									},
									Values: []ast.Expr{
										&ast.BasicLit{
											Kind:  "number",
											Value: "10",
										},
									},
								},
								&ast.FuncDecl{
									Name: &ast.Ident{
										Name: "constructor",
									},
									Type: &ast.FuncType{
										Params: &ast.FieldList{
											List: []*ast.Field{
												{
													Name: &ast.Ident{
														Name: "a",
													},
												},
												{
													Name: &ast.Ident{
														Name: "b",
													},
												},
											},
										},
									},
									Body: &ast.BlockStmt{
										List: []ast.Stmt{
											&ast.AssignStmt{
												LHS: []ast.Expr{
													&ast.SelectorExpr{
														Expr: &ast.Ident{
															Name: "this",
														},
														Sel: &ast.Ident{
															Name: "a",
														},
													},
												},
												RHS: []ast.Expr{
													&ast.Ident{
														Name: "a",
													},
												},
											},
											&ast.AssignStmt{
												LHS: []ast.Expr{
													&ast.SelectorExpr{
														Expr: &ast.Ident{
															Name: "this",
														},
														Sel: &ast.Ident{
															Name: "b",
														},
													},
												},
												RHS: []ast.Expr{
													&ast.Ident{
														Name: "b",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	runner(t, testcases)
}

func TestParseTopLevelExpressions(t *testing.T) {
	testcases := []testcase{
		{
			name: "ExpressionStatement",
			src:  "console.log('testing');",
			file: ast.File{
				Name: &ast.Ident{},
				Exprs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							Expr: &ast.Ident{
								Name: "console",
							},
							Sel: &ast.Ident{
								Name: "log",
							},
						},
						Args: []ast.Expr{
							&ast.BasicLit{
								Kind:  "string",
								Value: "testing",
							},
						},
					},
				},
			},
		},
	}

	runner(t, testcases)
}

func TestParseFunctionDeclaration(t *testing.T) {
	testcases := []testcase{
		{
			name: "SimpleFunction",
			src:  `function foo(a,b) {}`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.FuncDecl{
						Name: &ast.Ident{
							Name: "foo",
						},
						Type: &ast.FuncType{
							Params: &ast.FieldList{
								List: []*ast.Field{
									{
										Name: &ast.Ident{
											Name: "a",
										},
									},
									{
										Name: &ast.Ident{
											Name: "b",
										},
									},
								},
							},
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{},
						},
					},
				},
			},
		},
		{
			name: "ArrowFunction",
			src:  `const foo = (a,b) => {}`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.FuncDecl{
						Name: &ast.Ident{
							Name: "foo",
						},
						Type: &ast.FuncType{
							Params: &ast.FieldList{
								List: []*ast.Field{
									{
										Name: &ast.Ident{
											Name: "a",
										},
									},
									{
										Name: &ast.Ident{
											Name: "b",
										},
									},
								},
							},
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{},
						},
					},
				},
			},
		},
		{
			name: "SimpleBody",
			src: `
function f(a, b) {
	const c = a + b;
	console.log(c);
	const foo = doFoo(c);
	return foo;
}
			`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.FuncDecl{
						Name: &ast.Ident{
							Name: "f",
						},
						Type: &ast.FuncType{
							Params: &ast.FieldList{
								List: []*ast.Field{
									{
										Name: &ast.Ident{
											Name: "a",
										},
									},
									{
										Name: &ast.Ident{
											Name: "b",
										},
									},
								},
							},
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.AssignStmt{
									LHS: []ast.Expr{
										&ast.Ident{
											Name: "c",
										},
									},
									RHS: []ast.Expr{
										&ast.BinaryExpr{
											Left: &ast.Ident{
												Name: "a",
											},
											// Op: "+",
											Right: &ast.Ident{
												Name: "b",
											},
										},
									},
								},
								&ast.ExprStmt{
									Expr: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											Expr: &ast.Ident{
												Name: "console",
											},
											Sel: &ast.Ident{
												Name: "log",
											},
										},
										Args: []ast.Expr{
											&ast.Ident{
												Name: "c",
											},
										},
									},
								},
								&ast.AssignStmt{
									LHS: []ast.Expr{
										&ast.Ident{
											Name: "foo",
										},
									},
									RHS: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.Ident{
												Name: "doFoo",
											},
											Args: []ast.Expr{
												&ast.Ident{
													Name: "c",
												},
											},
										},
									},
								},
								&ast.ReturnStmt{
									Results: []ast.Expr{
										&ast.Ident{
											Name: "foo",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ComplexBody",
			src: `
function f() {
	const sum = (a,b) => {
		return a + b;
	};
	const sub = function(a,b) {
		return a - b;
	}
	const result = doSomething(sum(10,20));
	console.log(` + "`result: ${result}`);" + `
}
			`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.FuncDecl{
						Name: &ast.Ident{
							Name: "f",
						},
						Type: &ast.FuncType{
							Params:  &ast.FieldList{},
							Results: nil,
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.AssignStmt{
									LHS: []ast.Expr{
										&ast.Ident{
											Name: "sum",
										},
									},
									RHS: []ast.Expr{
										&ast.FuncLit{
											Type: &ast.FuncType{
												Results: nil,
												Params: &ast.FieldList{
													List: []*ast.Field{
														{
															Name: &ast.Ident{
																Name: "a",
															},
														},
														{
															Name: &ast.Ident{
																Name: "b",
															},
														},
													},
												},
											},
											Body: &ast.BlockStmt{
												List: []ast.Stmt{
													&ast.ReturnStmt{
														Results: []ast.Expr{
															&ast.BinaryExpr{
																Left: &ast.Ident{
																	Name: "a",
																},
																// Op: "+",
																Right: &ast.Ident{
																	Name: "b",
																},
															},
														},
													},
												},
											},
										},
									},
								},
								&ast.AssignStmt{
									LHS: []ast.Expr{
										&ast.Ident{
											Name: "sub",
										},
									},
									RHS: []ast.Expr{
										&ast.FuncLit{
											Type: &ast.FuncType{
												Results: nil,
												Params: &ast.FieldList{
													List: []*ast.Field{
														{
															Name: &ast.Ident{
																Name: "a",
															},
														},
														{
															Name: &ast.Ident{
																Name: "b",
															},
														},
													},
												},
											},
											Body: &ast.BlockStmt{
												List: []ast.Stmt{
													&ast.ReturnStmt{
														Results: []ast.Expr{
															&ast.BinaryExpr{
																Left: &ast.Ident{
																	Name: "a",
																},
																// Op: "-",
																Right: &ast.Ident{
																	Name: "b",
																},
															},
														},
													},
												},
											},
										},
									},
								},
								&ast.AssignStmt{
									LHS: []ast.Expr{
										&ast.Ident{
											Name: "result",
										},
									},
									RHS: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.Ident{
												Name: "doSomething",
											},
											Args: []ast.Expr{
												&ast.CallExpr{
													Fun: &ast.Ident{
														Name: "sum",
													},
													Args: []ast.Expr{
														&ast.BasicLit{
															Kind:  "number",
															Value: "10",
														},
														&ast.BasicLit{
															Kind:  "number",
															Value: "20",
														},
													},
												},
											},
										},
									},
								},
								&ast.ExprStmt{
									Expr: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											Expr: &ast.Ident{
												Name: "console",
											},
											Sel: &ast.Ident{
												Name: "log",
											},
										},
										Args: []ast.Expr{
											&ast.TemplateExpr{
												Value: "result: ${result}",
												Subs: []ast.Expr{
													&ast.Ident{
														Name: "result",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	runner(t, testcases)
}

func TestParseVariableDeclaration(t *testing.T) {
	testcases := []testcase{
		{
			name: "VarDeclaration",
			src:  `var foo = "bar"`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ValueDecl{
						Names: []*ast.Ident{
							{
								Name: "foo",
							},
						},
						Values: []ast.Expr{
							&ast.BasicLit{
								Kind:  "string",
								Value: "bar",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstDeclaration",
			src:  `const foo = 123`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ValueDecl{
						Names: []*ast.Ident{
							{
								Name: "foo",
							},
						},
						Values: []ast.Expr{
							&ast.BasicLit{
								Kind:  "number",
								Value: "123",
							},
						},
					},
				},
			},
		},
		{
			name: "LetDeclaration",
			src:  `let foo = {bar: 1.10, obj: {other: foo} }`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ValueDecl{
						Names: []*ast.Ident{
							{
								Name: "foo",
							},
						},
						Values: []ast.Expr{
							&ast.ObjectExpr{
								Elts: []ast.Expr{
									&ast.KeyValueExpr{
										Key: &ast.Ident{
											Name: "bar",
										},
										Value: &ast.BasicLit{
											Kind:  "number",
											Value: "1.10",
										},
									},
									&ast.KeyValueExpr{
										Key: &ast.Ident{
											Name: "obj",
										},
										Value: &ast.ObjectExpr{
											Elts: []ast.Expr{
												&ast.KeyValueExpr{
													Key:   &ast.Ident{Name: "other"},
													Value: &ast.Ident{Name: "foo"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ObjectCreation",
			src:  `let foo = new Object(1, "foo", bar, {testing: true})`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ValueDecl{
						Names: []*ast.Ident{
							{
								Name: "foo",
							},
						},
						Values: []ast.Expr{
							&ast.ObjectExpr{
								Name: &ast.Ident{
									Name: "foo",
								},
								Type: &ast.Ident{
									Name: "Object",
								},
								Elts: []ast.Expr{
									&ast.BasicLit{
										Kind:  "number",
										Value: "1",
									},
									&ast.BasicLit{
										Kind:  "string",
										Value: "foo",
									},
									&ast.Ident{
										Name: "bar",
									},
									&ast.ObjectExpr{
										Name: nil,
										Type: nil,
										Elts: []ast.Expr{
											&ast.KeyValueExpr{
												Key: &ast.Ident{
													Name: "testing",
												},
												Value: &ast.BasicLit{
													Kind:  "boolean",
													Value: "true",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ArrayCreation",
			src:  `let a = ['foo', 10, foo]`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ValueDecl{
						Names: []*ast.Ident{
							{
								Name: "a",
							},
						},
						Values: []ast.Expr{
							&ast.ObjectExpr{
								Elts: []ast.Expr{
									&ast.BasicLit{
										Kind:  "string",
										Value: "foo",
									},
									&ast.BasicLit{
										Kind:  "number",
										Value: "10",
									},
									&ast.Ident{
										Name: "foo",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "BinaryExpression",
			src:  "const a = 10 + 10 * (20 / x)",
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ValueDecl{
						Names: []*ast.Ident{
							{
								Name: "a",
							},
						},
						Values: []ast.Expr{
							&ast.BinaryExpr{
								Left: &ast.BasicLit{
									Kind:  "number",
									Value: "10",
								},
								// Op: "+",
								Right: &ast.BinaryExpr{
									Left: &ast.BasicLit{
										Kind:  "number",
										Value: "10",
									},
									// Op: "*",
									Right: &ast.BinaryExpr{
										Left: &ast.BasicLit{
											Kind:  "number",
											Value: "20",
										},
										// Op: "/",
										Right: &ast.Ident{
											Name: "x",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "DefaultParameters",
			src:  `function f(a, b = 1) {}`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.FuncDecl{
						Name: &ast.Ident{
							Name: "f",
						},
						Type: &ast.FuncType{
							Params: &ast.FieldList{
								List: []*ast.Field{
									{
										Name: &ast.Ident{
											Name: "a",
										},
									},
									{
										Name: &ast.ObjectExpr{
											Name: &ast.Ident{
												Name: "b",
											},
											Elts: []ast.Expr{
												&ast.BasicLit{
													Kind:  "number",
													Value: "1",
												},
											},
										},
									},
								},
							},
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{},
						},
					},
				},
			},
		},
		{
			name: "MultiVariableAssingment",
			src:  "const a = 1, b = 2;",
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ValueDecl{
						Names: []*ast.Ident{
							{
								Name: "a",
							},
							{
								Name: "b",
							},
						},
						Values: []ast.Expr{
							&ast.BasicLit{
								Kind:  "number",
								Value: "1",
							},
							&ast.BasicLit{
								Kind:  "number",
								Value: "2",
							},
						},
					},
				},
			},
		},
		{
			name: "VariableDeclarationWithoutInitValue",
			src:  "let foo;",
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ValueDecl{
						Names: []*ast.Ident{
							{
								Name: "foo",
							},
						},
					},
				},
			},
		},
	}
	runner(t, testcases)
}

func TestParseFileImports(t *testing.T) {
	testcases := []testcase{
		{
			name: "EntireModule",
			src:  "import * as foo from 'bar'",
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ImportDecl{
						Name: &ast.Ident{
							Name: "foo",
						},
						Path: &ast.Ident{
							Name: "bar",
						},
					},
				},
			},
		},
		{
			name: "DefaultExportFromModule",
			src:  "import foo from 'bar'",
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ImportDecl{
						Name: &ast.Ident{
							Name: "foo",
						},
						Path: &ast.Ident{
							Name: "bar",
						},
					},
				},
			},
		},
		{
			name: "SingleExportFromModule",
			src:  `import { foo } from 'bar';`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ImportDecl{
						Name: &ast.Ident{
							Name: "foo",
						},
						Path: &ast.Ident{
							Name: "bar",
						},
					},
				},
			},
		},
		{
			name: "MultipleExportFromModule",
			src:  `import { foo, bar } from 'baz';`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ImportDecl{
						Name: &ast.Ident{
							Name: "foo",
						},
						Path: &ast.Ident{
							Name: "baz",
						},
					},
					&ast.ImportDecl{
						Name: &ast.Ident{
							Name: "bar",
						},
						Path: &ast.Ident{
							Name: "baz",
						},
					},
				},
			},
		},
		{
			name: "SideEffect",
			src:  `import 'foo'`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ImportDecl{
						Path: &ast.Ident{
							Name: "foo",
						},
					},
				},
			},
		},
		{
			name: "Aliases",
			src:  `import { foo as baz } from 'bar';`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ImportDecl{
						Name: &ast.Ident{
							Name: "foo",
						},
						Path: &ast.Ident{
							Name: "bar",
						},
					},
				},
			},
		},
		{
			name: "Require",
			src:  `const foo = require('foo')`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ImportDecl{
						Name: &ast.Ident{
							Name: "foo",
						},
						Path: &ast.Ident{
							Name: "foo",
						},
					},
				},
			},
		},
		{
			name: "RequireUsingDifferentName",
			src:  `const foo = require('bar')`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ImportDecl{
						Name: &ast.Ident{
							Name: "foo",
						},
						Path: &ast.Ident{
							Name: "bar",
						},
					},
				},
			},
		},
		{
			name: "RequireAliases",
			src:  `const foo = require('bar')`,
			file: ast.File{
				Name: &ast.Ident{},
				Decls: []ast.Decl{
					&ast.ImportDecl{
						Name: &ast.Ident{
							Name: "foo",
						},
						Path: &ast.Ident{
							Name: "bar",
						},
					},
				},
			},
		},
	}
	runner(t, testcases)
}

func runner(t *testing.T, testcases []testcase) {
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			file, err := javascript.ParseFile("", []byte(tt.src))
			require.Nil(t, err, "Expected nil error to parse source")
			assert.Equal(t, tt.file, *file)
		})
	}
}
