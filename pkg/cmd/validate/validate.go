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

package validate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tektoncd/catlin/pkg/app"
	"github.com/tektoncd/catlin/pkg/parser"
	"github.com/tektoncd/catlin/pkg/validator"
)

var cat = []string{}

func getFilesInDir(path string) ([]os.FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	files, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	if info.IsDir() {

		files, err := getFilesInDir(path)
		if err != nil {
			return false
		}

		for _, file := range files {
			if filepath.Ext(file.Name()) == ".yaml" {
				return true
			}
		}
		return false
	}

	return true
}

func validResourcePath() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires at least 1 path to tekton resource yaml but received none")
		}

		for _, filePath := range args {
			if !fileExists(filePath) {
				return fmt.Errorf("valid Tekton resource not found in the path - %s", filePath)
			}
		}

		return nil
	}
}

func validateFlags(cmd *cobra.Command, args []string, versioning string) error {
	if versioning != validator.DirectoryBasedVersioning && versioning != validator.GitBasedVersioning {
		return fmt.Errorf("invalid versioning: %s, expecting git or directory", versioning)
	}

	return nil
}

func Command(cli app.CLI) *cobra.Command {
	var versioning string
	cmd := &cobra.Command{
		Use:     "validate",
		Aliases: []string{"verify"},
		Args:    validResourcePath(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return validateFlags(cmd, args, versioning)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return validateResources(cli, args, versioning)
		},
	}

	cmd.PersistentFlags().StringVar(&versioning, "versioning", "directory", "Versioning type of catalog (directory/git)")

	return cmd
}

func validateResources(cli app.CLI, args []string, versioning string) error {

	out := cli.Stream().Out
	for _, filePath := range args {
		info, _ := os.Stat(filePath)
		if info.IsDir() {

			files, err := getFilesInDir(filePath)
			if err != nil {
				return errors.New(err.Error())
			}

			for _, file := range files {
				if filepath.Ext(file.Name()) == ".yaml" {
					fileWithPath := filePath
					if strings.HasSuffix(fileWithPath, "/") {
						fileWithPath = fileWithPath + file.Name()
					} else {
						fileWithPath = fileWithPath + "/" + file.Name()
					}
					_, err := fmt.Fprintf(out, "FILE: %s\n", fileWithPath)
					if err != nil {
						return err
					}
					err = validate(cli, fileWithPath, versioning)
					if err != nil {
						return err
					}
				}
			}
		} else if filepath.Ext(filePath) == ".yaml" {
			_, err := fmt.Fprintf(out, "FILE: %s\n", filePath)
			if err != nil {
				return err
			}
			err = validate(cli, filePath, versioning)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func validate(cli app.CLI, path, versioning string) error {

	r, err := os.Open(path)
	if err != nil {
		return err
	}

	// parse
	parser := parser.ForReader(r)
	res, err := parser.Parse()
	if err != nil {
		return err
	}

	if len(cat) == 0 {
		// Get the predefined category list
		cat, err = validator.GetCategories()
		if err != nil {
			return err
		}
	}

	// run validators
	validators := []validator.Validator{
		validator.NewPathValidator(res, path, versioning),
		validator.NewContentValidator(res, cat),
		validator.ForKind(res),
	}

	result := validator.Result{}
	for _, v := range validators {
		result.Append(v.Validate())
	}

	// print result
	out := cli.Stream().Out
	for _, v := range result.Lints {
		switch v.Kind {
		case validator.Error:
			_, err := fmt.Fprintf(out, "ERROR: %s\n", v.Message)
			if err != nil {
				return err
			}
		case validator.Warning:
			_, err := fmt.Fprintf(out, "WARN : %s\n", v.Message)
			if err != nil {
				return err
			}
		case validator.Recommendation:
			_, err := fmt.Fprintf(out, "HINT : %s\n", v.Message)
			if err != nil {
				return err
			}
		case validator.Info:
			_, err := fmt.Fprintf(out, "INFO : %s\n", v.Message)
			if err != nil {
				return err
			}
		default:
			level := strings.ToUpper(fmt.Sprint(v.Kind))
			_, err := fmt.Fprintf(out, "%s : %s\n", level, v.Message)
			if err != nil {
				return err
			}
		}
	}
	if result.Errors != 0 {
		return fmt.Errorf("%s failed validation", path)
	}

	return nil
}
