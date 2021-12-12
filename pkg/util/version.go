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

package util

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Main     int
	Sub      int
	Revision int
}

type ComparisonLevel int

const (
	MainVersion     ComparisonLevel = 0
	SubVersion      ComparisonLevel = 1
	RevisionVersion ComparisonLevel = 2
)

func NewVersion(ver string) (*Version, error) {
	if !strings.HasPrefix(ver, "v") {
		return nil, fmt.Errorf("wrong format: %s", ver)
	}
	vs := strings.Split(strings.TrimPrefix(ver, "v"), ".")
	ret := &Version{Main: 0, Sub: 0, Revision: 0}
	for i, v := range vs {
		vInt, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("wrong format version: %d/%s/%s", i, v, ver)
		}
		switch i {
		case 0:
			ret.Main = vInt
		case 1:
			ret.Sub = vInt
		case 2:
			ret.Revision = vInt
		default:
			return nil, fmt.Errorf("wrong format: %s", ver)
		}
	}
	return ret, nil
}

func (v *Version) Compare(ver *Version, lvl ComparisonLevel) int {
	cmpFunc := func(ahead bool, eq bool) int {
		if ahead {
			return 1
		}
		if eq {
			return 0
		}
		return -1
	}
	if v.Main != ver.Main || lvl == MainVersion {
		return cmpFunc(v.Main > ver.Main, v.Main == ver.Main)
	}
	if v.Sub != ver.Sub || lvl == SubVersion {
		return cmpFunc(v.Sub > ver.Sub, v.Sub == ver.Sub)
	}
	if v.Revision != ver.Revision || lvl == RevisionVersion {
		return cmpFunc(v.Revision > ver.Revision, v.Revision == ver.Revision)
	}
	return 0
}

func CheckRegisterPluginTkeelVersion(dependVersion string, currVersion string) (bool, error) {
	dVer, err := NewVersion(dependVersion)
	if err != nil {
		return false, fmt.Errorf("error depend version: %w", err)
	}
	cVer, err := NewVersion(currVersion)
	if err != nil {
		return false, fmt.Errorf("error current version: %w", err)
	}
	if cVer.Compare(dVer, SubVersion) < 0 {
		return false, nil
	}
	return true, nil
}
