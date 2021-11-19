/*
Copyright 2021 The tKeel Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
	http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tkeel-io/tkeel/pkg/version"
)

// VersionCmd represents the version command.
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of metadata service",
	Long:  `All software has versions. This is metadata service`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s", version.Version)
		fmt.Printf("Build Date: %s", version.BuildDate)
		fmt.Printf("Git Commit: %s", version.GitCommit)
		fmt.Printf("Git Version: %s", version.GitVersion)
		fmt.Printf("Go Version: %s", version.GoVersion)
		fmt.Printf("OS / Arch: %s", version.OsArch)
	},
}
