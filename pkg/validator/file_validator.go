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
	"path/filepath"
	"strings"

	"github.com/tektoncd/catlin/pkg/parser"
)

const (
	GitBasedVersioning       = "git"
	DirectoryBasedVersioning = "directory"
)

type PathValidator struct {
	path       string
	versioning string
	res        *parser.Resource
}

var _ Validator = (*PathValidator)(nil)

func NewPathValidator(r *parser.Resource, path, versioning string) *PathValidator {
	return &PathValidator{path: path, versioning: versioning, res: r}
}

func (v *PathValidator) Validate() Result {
	r := v.res
	name := r.Name
	kind := strings.ToLower(r.Kind)
	version := r.Version()

	result := Result{}

	absPath, err := filepath.Abs(v.path)
	if err != nil {
		result.Error("invalid path: %s - %s", v.path, err)
		return result
	}

	var expectedPath string
	switch v.versioning {
	case GitBasedVersioning:
		expectedPath = filepath.Join(kind, name, name+".yaml")
	case DirectoryBasedVersioning:
		expectedPath = filepath.Join(kind, name, version, name+".yaml")
	default:
		result.Error("invalid versioning, expecting git or directory, but got: %s", v.versioning)
		return result
	}

	if !strings.HasSuffix(absPath, expectedPath) {
		result.Error("Resource path is invalid; expected path: %s", expectedPath)
	}
	return result
}
