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
	"os"
	"runtime"
)

// Version The tkeel platform Version.
const _TKEEL_VERSION = "TKEEL_VERSION"

var Version string

// GitCommit The git commit that was compiled. This will be filled in by the compiler.
var (
	GitCommit string
	GitBranch string
)

// GitVersion The main version number that is being run at the moment.
var GitVersion string

// BuildDate The build datetime at the moment.
var BuildDate = ""

// GoVersion The go compiler version.
var GoVersion = runtime.Version()

// OsArch The system info.
var OsArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)

func init() {
	if ver := os.Getenv(_TKEEL_VERSION); ver != "" {
		Version = ver
	}
}
