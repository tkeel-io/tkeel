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

package version

import (
	"fmt"
	"runtime"
)

var (
	// version The tkeel platform Version.
	version string = "v0.2.0"
	// gitCommit The git commit that was compiled. This will be filled in by the compiler.
	gitCommit string
	// gitTreeState is the state of the git tree
	gitTreeState string
	// builtAt The build datetime at the moment.
	builtAt string
	// metadata is extra build time data
	metadata = ""
)

type BuildInfo struct {
	// Version is the current semver.
	Version string `json:"version,omitempty"`
	// GitCommit is the git sha1.
	GitCommit string `json:"git_commit,omitempty"`
	// GitTreeState is the state of the git tree.
	GitTreeState string `json:"git_tree_state,omitempty"`
	// BuiltAt is the time when this program built.
	BuiltAt string `json:"built_at,omitempty"`
	// GoVersion is the version of the Go compiler used.
	GoVersion string `json:"go_version,omitempty"`
	// OSArch is what architecture the program is built using.
	OSArch string `json:"os_arch,omitempty"`
}

func Get() BuildInfo {
	return BuildInfo{
		Version:      GetVersion(),
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		BuiltAt:      builtAt,
		GoVersion:    runtime.Version(),
		OSArch:       fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH),
	}
}

// GetVersion returns the semver string of the version
func GetVersion() string {
	if metadata == "" {
		return version
	}
	return version + "+" + metadata
}