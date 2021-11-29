package helm

import (
	"bytes"
	"context"
	"regexp"

	"github.com/pkg/errors"
	"github.com/tkeel-io/tkeel/pkg/output"
)

const versionRegex = `^\d+\.\d+.\d+$`

var ErrVersionPattern = errors.New("invalid version")

func AddRepo(addr string) error {
	return addRepo(privateRepoName, addr)
}

func DeleteOwnRepo() error {
	return deleteRepo(privateRepoName)
}

func ListRepo(format string) ([]byte, error) {
	o, err := output.ParseFormat(format)
	if err != nil {
		err = errors.Wrap(err, "parse format err")
		return nil, err
	}

	data, err := listRepo()
	if err != nil {
		return nil, err
	}

	var listbuf bytes.Buffer
	if err := data.encodeByFormat(&listbuf, o); err != nil {
		return nil, err
	}

	return listbuf.Bytes(), nil
}

func ListInstallable(format string, updateRepo bool) ([]byte, error) {
	o, err := output.ParseFormat(format)
	if err != nil {
		err = errors.Wrap(err, "parse format err")
		return nil, err
	}
	if updateRepo {
		if err = RepoUpdate(); err != nil {
			return nil, errors.Wrap(err, "update repo failed")
		}
	}

	pluginWriter, err := searchAll()
	if err != nil {
		return nil, errors.Wrap(err, "search helm repo failed")
	}
	buf := new(bytes.Buffer)
	if err := output.Write(buf, o, pluginWriter); err != nil {
		err = errors.Wrap(err, "convert to output err")
		return nil, err
	}
	return buf.Bytes(), nil
}

func ListInstalled(format string) ([]byte, error) {
	o, err := output.ParseFormat(format)
	if err != nil {
		err = errors.Wrap(err, "parse format err")
		return nil, err
	}
	listWriter, err := list()
	if err != nil {
		return nil, errors.Wrap(err, "search helm repo failed")
	}
	buf := new(bytes.Buffer)
	if err := output.Write(buf, o, listWriter); err != nil {
		err = errors.Wrap(err, "convert to output err")
		return nil, err
	}
	return buf.Bytes(), nil
}

func Install(ctx context.Context, name, chart, version string) error {
	if version != "" {
		if version == "latest" {
			version = ""
			goto install
		}
		matched, err := regexp.MatchString(versionRegex, version)
		if err != nil {
			err = errors.Wrap(err, "check regexp err")
			return err
		}
		if !matched {
			return ErrVersionPattern
		}
	}
install:
	return installChart(name, chart, version, loadComponentChart())
}

func Uninstall(ctx context.Context, name ...string) error {
	return uninstallChart(name...)
}
