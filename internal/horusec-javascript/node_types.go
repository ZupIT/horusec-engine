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

const (
	// ------------------------------------------------
	//
	//
	// Declaration nodes

	ClassDeclaration             = "class_declaration"
	FunctionDeclaration          = "function_declaration"
	GeneratorFunctionDeclaration = "generation_function_declaration"
	LexicalDeclaration           = "lexical_declaration"
	VariableDeclaration          = "variable_declaration"
	VariableDeclarator           = "variable_declarator"
	MethodDefinition             = "method_definition"
	PublicFieldDefinition        = "public_field_definition"

	// ------------------------------------------------
	//
	//
	// Expression nodes

	AssignmentExpression    = "assignment_expression"
	BinaryExpression        = "binary_expression"
	CallExpression          = "call_expression"
	MemberExpression        = "member_expression"
	Identifier              = "identifier"
	This                    = "this"
	PropertyIdentifier      = "property_identifier"
	String                  = "string"
	Number                  = "number"
	Object                  = "object"
	True                    = "true"
	False                   = "false"
	Pair                    = "pair"
	ArrowFunction           = "arrow_function"
	NewExpression           = "new_expression"
	Array                   = "array"
	ParenthesizedExpression = "parenthesized_expression"
	AssignmentPattern       = "assignment_pattern"
	SequenceExpression      = "sequence_expression"
	TemplateString          = "template_string"
	Function                = "function"
	UpdateExpression        = "update_expression"

	// ------------------------------------------------
	//
	//
	// Statement nodes

	ExpressionStatement = "expression_statement"
	ImportStatement     = "import_statement"
	NamedImports        = "named_imports"
	NamespaceImport     = "namespace_import"
	ReturnStatement     = "return_statement"
	IfStatement         = "if_statement"
	ElseClause          = "else_clause"
	StatementBlock      = "statement_block"
	TryStatement        = "try_statement"
	FinallyClause       = "finally_clause"
	CatchClause         = "catch_clause"
	WhileStatement      = "while_statement"
	SwitchStatement     = "switch_statement"
	SwitchBody          = "switch_body"
	SwitchCase          = "switch_case"
	SwitchDefault       = "switch_default"
	BreakStatement      = "break_statement"
	ForStatement        = "for_statement"
	ForInStatement      = "for_in_statement"

	// ------------------------------------------------
	//
	//
	// Others nodes

	Comment = "comment"
)
