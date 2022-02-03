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
	"io"
	"sort"
)

func (f *File) String() string {
	return "file " + f.name
}

// nolint: funlen // To many cases to handle.
func (s *Signature) String() string {
	buf := bytes.NewBufferString("(")

	params := make([]Value, 0, len(s.Params))
	for _, p := range s.Params {
		params = append(params, p)
	}
	joinValues(buf, params)

	if len(s.Results) > 0 {
		results := make([]Value, 0, len(s.Results))
		for _, r := range s.Results {
			results = append(results, r)
		}
		buf.WriteString("-> ")
		joinValues(buf, results)
	}

	buf.WriteString(")")

	return buf.String()
}

func (c *Call) String() string {
	buf := bytes.NewBufferString(c.Function.Name())

	buf.WriteString("(")
	joinValues(buf, c.Args)
	buf.WriteString(")")

	return buf.String()
}

// WriteTo writes to w a human-readable summary of file.
func (f *File) WriteTo(w io.Writer) (int64, error) {
	buf := bytes.NewBufferString("")
	WriteFile(buf, f)
	n, err := w.Write(buf.Bytes())

	return int64(n), err
}

// WriteFile writes to buf a human-readable summary of f.
//
// nolint: funlen,gocyclo // To many cases to handle.
func WriteFile(buf *bytes.Buffer, f *File) {
	fmt.Fprintf(buf, "%s:\n", f)

	var names []string
	maxname := 0
	for name := range f.Members {
		if l := len(name); l > maxname {
			maxname = l
		}
		names = append(names, name)
	}
	sort.Strings(names)

	// Write all imports before.
	for _, mem := range f.Members {
		if ext, ok := mem.(*ExternalMember); ok {
			fmt.Fprintf(buf, "  import  %-*s\n", maxname, ext.Path)
		}
	}

	for _, name := range names {
		switch mem := f.Members[name].(type) {
		case *ExternalMember:
		case *Function:
			fmt.Fprintf(buf, "  func    %-*s %s\n", maxname, name, mem.Signature)
		default:
			panic(fmt.Sprintf("ir.WriteFile: unhandled member type: %T", mem))
		}
	}

	fmt.Fprintf(buf, "\n")
}

// WriteTo writes to w a human-readable "disassembly" of function.
func (fn *Function) WriteTo(w io.Writer) (int64, error) {
	buf := bytes.NewBufferString("")
	WriteFunction(buf, fn)
	n, err := w.Write(buf.Bytes())

	return int64(n), err
}

// WriteFunction writes to buf a human-readable "disassembly" of fn.
//
// nolint: funlen,gocyclo // To many cases to handle.
func WriteFunction(buf *bytes.Buffer, fn *Function) {
	fmt.Fprintf(buf, "# Name: %s\n", fn.Name())

	if fn.File != nil {
		fmt.Fprintf(buf, "# File: %s\n", fn.File.Name())
	}

	fmt.Fprintf(buf, "# Location: %s:%d:%d\n", fn.File.Name(), fn.syntax.Start().Row, fn.syntax.Start().Column)

	fmt.Fprintf(buf, "func %s%s:\n", fn.Name(), fn.Signature)

	if fn.Blocks == nil {
		buf.WriteString("\t(external)\n")
	}

	const width = 80
	for _, b := range fn.Blocks {
		if b == nil {
			fmt.Fprintf(buf, ".nil:\n")

			continue
		}

		// Pretty write the index and name of the current block using indentation.
		n, _ := fmt.Fprintf(buf, "%d:", b.Index)
		fmt.Fprintf(buf, "%*s\n", width-n-len(b.Comment), b.Comment)

		for _, instr := range b.Instrs {
			buf.WriteString("\t")
			switch instr := instr.(type) {
			case *Call:
				buf.WriteString(instr.String())
			default:
				panic(fmt.Sprintf("ir.WriteFunction: unhandled instruction type: %T", instr))
			}
			buf.WriteString("\n")
		}
	}

	fmt.Fprintf(buf, "\n")
}

// joinValues concatenates the values on buf. A comma separator string is placed between elements.
func joinValues(buf *bytes.Buffer, values []Value) {
	for i, value := range values {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(value.Name())
	}
}
