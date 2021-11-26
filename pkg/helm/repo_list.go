package helm

import (
	"io"

	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/tkeel-io/tkeel/pkg/output"
	"helm.sh/helm/v3/pkg/repo"
)

func listRepo() (*repoListWriter, error) {
	b, err := getRepositoryFormDapr()
	if err != nil {
		err = errors.Wrap(err, "failed try to get repository.yaml config")
		return nil, err
	}
	f, err := newHelmRepoFile(b)
	if err != nil {
		err = errors.Wrap(err, "new helm repo.File err")
		return nil, err
	}
	if len(f.Repositories) == 0 {
		return nil, errors.New("no repositories to show")
	}

	return &repoListWriter{f.Repositories}, nil
}

type repoListWriter struct {
	repos []*repo.Entry
}

func (r *repoListWriter) WriteJSON(out io.Writer) error {
	return r.encodeByFormat(out, output.JSON)
}

func (r *repoListWriter) WriteYAML(out io.Writer) error {
	return r.encodeByFormat(out, output.YAML)
}

func (r repoListWriter) WriteTable(out io.Writer) error {
	table := uitable.New()
	table.AddRow("NAME", "URL")
	for _, re := range r.repos {
		table.AddRow(re.Name, re.URL)
	}
	if err := output.EncodeTable(out, table); err != nil {
		err = errors.Wrap(err, "encode data to table format err")
		return err
	}
	return nil
}

func (r *repoListWriter) encodeByFormat(out io.Writer, format output.Format) error {
	// Initialize the array so no results returns an empty array instead of null.
	repolist := make([]repositoryElement, 0, len(r.repos))

	for _, re := range r.repos {
		repolist = append(repolist, repositoryElement{Name: re.Name, URL: re.URL})
	}

	switch format {
	case output.JSON:
		if err := output.EncodeJSON(out, repolist); err != nil {
			err = errors.Wrap(err, "encode data to json fomat err")
			return err
		}
		return nil
	case output.YAML:
		if err := output.EncodeYAML(out, repolist); err != nil {
			err = errors.Wrap(err, "encode data to yaml format err")
			return err
		}
		return nil
	case output.TABLE:
		return r.WriteTable(out)
	}

	// Because this is a non-exported function and only called internally by
	// WriteJSON and WriteYAML, we shouldn't get invalid types.
	return nil
}

type repositoryElement struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
