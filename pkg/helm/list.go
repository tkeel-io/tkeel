package helm

import (
	"fmt"
	"io"
	"strconv"

	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/tkeel-io/tkeel/pkg/output"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

func list() (*releaseListWriter, error) {
	listClient := action.NewList(defaultCfg)
	listClient.SetStateMask()
	results, err := listClient.Run()
	if err != nil {
		err = errors.Wrap(err, "run helm list err")
		return nil, err
	}

	return newReleaseListWriter(results, ""), nil
}

type releaseElement struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Revision   string `json:"revision"`
	Updated    string `json:"updated"`
	Status     string `json:"status"`
	Chart      string `json:"chart"`
	AppVersion string `json:"app_version"`
}

type releaseListWriter struct {
	releases []releaseElement
}

func newReleaseListWriter(releases []*release.Release, timeFormat string) *releaseListWriter {
	// Initialize the array so no results returns an empty array instead of null
	elements := make([]releaseElement, 0, len(releases))
	for _, r := range releases {
		element := releaseElement{
			Name:       r.Name,
			Namespace:  r.Namespace,
			Revision:   strconv.Itoa(r.Version),
			Status:     r.Info.Status.String(),
			Chart:      formatChartname(r.Chart),
			AppVersion: formatAppVersion(r.Chart),
		}

		t := "-"
		if tspb := r.Info.LastDeployed; !tspb.IsZero() {
			if timeFormat != "" {
				t = tspb.Format(timeFormat)
			} else {
				t = tspb.String()
			}
		}
		element.Updated = t

		elements = append(elements, element)
	}
	return &releaseListWriter{elements}
}

func (r *releaseListWriter) WriteTable(out io.Writer) error {
	table := uitable.New()
	table.AddRow("NAME", "NAMESPACE", "REVISION", "UPDATED", "STATUS", "CHART", "APP VERSION")
	for _, r := range r.releases {
		table.AddRow(r.Name, r.Namespace, r.Revision, r.Updated, r.Status, r.Chart, r.AppVersion)
	}
	return output.EncodeTable(out, table)
}

func (r *releaseListWriter) WriteJSON(out io.Writer) error {
	return r.encodeByFormat(out, output.JSON)
}

func (r *releaseListWriter) WriteYAML(out io.Writer) error {
	return r.encodeByFormat(out, output.YAML)
}

func (r *releaseListWriter) encodeByFormat(out io.Writer, format output.Format) error {
	// Initialize the array so no results returns an empty array instead of null.
	releases := make([]releaseElement, 0, len(r.releases))

	for _, re := range r.releases {
		releases = append(releases, releaseElement{
			Name:       re.Name,
			Namespace:  re.Namespace,
			Revision:   re.Revision,
			Updated:    re.Updated,
			Status:     re.Status,
			Chart:      re.Chart,
			AppVersion: re.AppVersion,
		})
	}

	switch format {
	case output.JSON:
		if err := output.EncodeJSON(out, releases); err != nil {
			err = errors.Wrap(err, "encode data to json fomat err")
			return err
		}
		return nil
	case output.YAML:
		if err := output.EncodeYAML(out, releases); err != nil {
			err = errors.Wrap(err, "encode data to yaml format err")
			return err
		}
		return nil
	case output.TABLE:
		return r.WriteTable(out)
	}

	return nil
}

func formatChartname(c *chart.Chart) string {
	if c == nil || c.Metadata == nil {
		// This is an edge case that has happened in prod, though we don't
		// know how: https://github.com/helm/helm/issues/1347
		return "MISSING"
	}
	return fmt.Sprintf("%s-%s", c.Name(), c.Metadata.Version)
}

func formatAppVersion(c *chart.Chart) string {
	if c == nil || c.Metadata == nil {
		// This is an edge case that has happened in prod, though we don't
		// know how: https://github.com/helm/helm/issues/1347
		return "MISSING"
	}
	return c.AppVersion()
}
