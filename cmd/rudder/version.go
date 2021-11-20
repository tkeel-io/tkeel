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

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tkeel-io/tkeel/pkg/version"
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of metadata service",
		Long:  `All software has versions. This is metadata service`,
		Run: func(cmd *cobra.Command, args []string) {
			info := version.Get()
			fmt.Println()
			fmt.Fprintf(os.Stdout, "Version: %s\n", info.Version)
			fmt.Fprintf(os.Stdout, "Build Date: %s\n", info.BuiltAt)
			fmt.Fprintf(os.Stdout, "Git Commit: %s\n", info.GitCommit)
			fmt.Fprintf(os.Stdout, "Git Tree State: %s\n", info.GitTreeState)
			fmt.Fprintf(os.Stdout, "Go Version: %s\n", info.GoVersion)
			fmt.Fprintf(os.Stdout, "OS / Arch: %s\n", info.OSArch)
		},
	}
}
