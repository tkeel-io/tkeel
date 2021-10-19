package keel

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
	if len(vs) != 3 {
		return nil, fmt.Errorf("wrong format: %s", ver)
	}
	m, err := strconv.Atoi(vs[0])
	if err != nil {
		return nil, fmt.Errorf("wrong format main: %s/%s", vs[0], ver)
	}
	s, err := strconv.Atoi(vs[1])
	if err != nil {
		return nil, fmt.Errorf("wrong format sub: %s/%s", vs[1], ver)
	}
	r, err := strconv.Atoi(vs[2])
	if err != nil {
		return nil, fmt.Errorf("wrong format revision: %s/%s", vs[2], ver)
	}
	return &Version{
		Main:     m,
		Sub:      s,
		Revision: r,
	}, nil
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

func CheckRegisterPluginTkeelVersion(dependVersion string, currVersion string) bool {
	dVer, err := NewVersion(dependVersion)
	if err != nil {
		log.Errorf("error depend version: %s", err)
		return false
	}
	cVer, err := NewVersion(currVersion)
	if err != nil {
		log.Errorf("error current version: %s", err)
		return false
	}
	if cVer.Compare(dVer, SubVersion) < 0 {
		return false
	}
	return true
}
