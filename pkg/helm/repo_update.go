package helm

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

func RepoUpdate(names ...string) error {
	rf, err := loadRepoFile()
	if err != nil {
		return errors.Wrap(err, "load repo file failed")
	}

	var repos []*repo.ChartRepository
	updateAllRepos := len(names) == 0

	if !updateAllRepos {
		// Fail early if the user specified an invalid repo to update
		if err := checkRequestedRepos(names, rf.Repositories); err != nil {
			return err
		}
	}

	for _, repocfg := range rf.Repositories {
		if updateAllRepos || isRepoRequested(repocfg.Name, names) {
			r, err := repo.NewChartRepository(repocfg, getter.All(env))
			if err != nil {
				err = errors.Wrap(err, "new chart repository err")
				return err
			}
			if env.RepositoryCache != "" {
				r.CachePath = env.RepositoryCache
			}
			repos = append(repos, r)
		}
	}

	var (
		wg           sync.WaitGroup
		repoFailList []string
	)
	for _, re := range repos {
		wg.Add(1)
		go func(re *repo.ChartRepository) {
			defer wg.Done()
			if _, err := re.DownloadIndexFile(); err != nil {
				log.Warn(fmt.Sprintf("...Unable to get an update from the %q chart repository (%s):\n\t%s\n", re.Config.Name, re.Config.URL, err))
				repoFailList = append(repoFailList, re.Config.URL)
			}
		}(re)
	}
	wg.Wait()

	if len(repoFailList) > 0 {
		return fmt.Errorf("failed to update the following repositories: %s", repoFailList)
	}

	return nil
}

func checkRequestedRepos(requestedRepos []string, validRepos []*repo.Entry) error {
	for _, requestedRepo := range requestedRepos {
		found := false
		for _, repo := range validRepos {
			if requestedRepo == repo.Name {
				found = true
				break
			}
		}
		if !found {
			return errors.Errorf("no repositories found matching '%s'.  Nothing will be updated", requestedRepo)
		}
	}
	return nil
}

func isRepoRequested(repoName string, requestedRepos []string) bool {
	for _, requestedRepo := range requestedRepos {
		if repoName == requestedRepo {
			return true
		}
	}
	return false
}
