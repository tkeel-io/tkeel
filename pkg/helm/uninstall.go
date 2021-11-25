package helm

import (
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
)

func uninstallChart(names ...string) error {
	uninstallClint := action.NewUninstall(defaultCfg)
	for _, name := range names {
		_, err := uninstallClint.Run(name)
		if err != nil {
			err = errors.Wrap(err, "call uninstall err")
			return err
		}
	}
	return nil
}
