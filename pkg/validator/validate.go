// Copyright © 2020 The Tekton Authors.
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

package validator

import (
	"github.com/tektoncd/plumbing/catlin/pkg/parser"
)

type Validator interface {
	Validate() Result
}

func ForKind(res *parser.Resource) Validator {
	switch res.Kind {
	case "Task":
		return NewTaskValidator(res)
	default:
		return &noopValidator{kind: res.Kind}
	}
}

type noopValidator struct {
	kind string
}

var _ Validator = (*noopValidator)(nil)

func (v *noopValidator) Validate() Result {
	r := Result{}
	r.Info("no validator specific to kind %s", v.kind)
	return r
}
