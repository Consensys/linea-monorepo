// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package field

// Kind identifies which field a column, expression, or DAG node lives in.
type Kind uint8

const (
	Base Kind = iota
	Ext
)

func (k Kind) String() string {
	switch k {
	case Base:
		return "base"
	case Ext:
		return "ext"
	default:
		return "unknown"
	}
}

// Join returns the field needed to contain values from both inputs.
func Join(a, b Kind) Kind {
	if a == Ext || b == Ext {
		return Ext
	}
	return Base
}

// JoinAll returns the field needed to contain every input.
func JoinAll(fields ...Kind) Kind {
	res := Base
	for _, f := range fields {
		res = Join(res, f)
	}
	return res
}
