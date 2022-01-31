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
)

func (c *Call) String() string {
	buf := bytes.NewBufferString(c.Function.Name)
	buf.WriteString("(")

	for i, arg := range c.Args {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(arg.Name())
	}

	buf.WriteString(")")

	return buf.String()
}

func (c *Const) Name() string { return c.Value }
func (v *Var) Name() string   { return v.name }
