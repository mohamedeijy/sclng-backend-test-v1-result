package handle

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Scalingo/go-utils/errors/v2"
	"github.com/Scalingo/sclng-backend-test-v1/cache"
	"github.com/Scalingo/sclng-backend-test-v1/structs"
	"github.com/Scalingo/sclng-backend-test-v1/utils"
	"github.com/google/go-github/v66/github"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type ReposHandler struct {
	client   *github.Client
	numRepos int
	cache    *cache.Cache // valid for a set period of time and reloaded at expiration
	log      logrus.FieldLogger
}

// InitReposHandler initialize a ReposHandler struct
func InitReposHandler(log logrus.FieldLogger, numRepos int) (*ReposHandler, error) {
	log.Info("Initializing repos handler...")
	// server failure if not authenticated (unauthenticated rate limit makes this API virtually impossible)
	authToken, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		log.Fatal("Personal Github token not found")
		return nil, http.ErrAbortHandler
	}
	client := github.NewClient(nil).WithAuthToken(authToken)
	log.Info("Authenticated with GitHub personal token")

	rh := &ReposHandler{
		client:   client,
		numRepos: numRepos,
		cache:    cache.NewCache(10 * time.Minute), // we chose to invalidate the cache every 10mns
		log:      log,
	}

	// set cache data on launch before making the server available
	githubRepos, err := rh.fetchRepos()
	if err != nil {
		rh.log.Infof("Fail to fetch repos to instantiate cache instance : %v", err)
	} else {
		rh.SetCache(githubRepos)
	}
	return rh, nil
}

func (rh *ReposHandler) SetCache(data []*structs.GithubRepo) {
	rh.cache.Set(data)
	rh.log.Info("Cache instantiated")
}

// ServeHTTP implement Handler interface
func (rh *ReposHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var githubRepos []*structs.GithubRepo
	var err error

	// check if cache valid call GitHub API if not and set cache again
	cacheInstance, ok := rh.cache.GetCacheData()
	if ok {
		rh.log.Info("Retrieved cache instance")
		githubRepos = cacheInstance
	} else {
		rh.log.Info("Cache invalid, fetching from GitHub API...")
		githubRepos, err = rh.fetchRepos()
		if err != nil {
			return err
		}
		rh.cache.Set(githubRepos)
	}

	// build up filter function by parsing the query parameters
	filters, err := utils.ParseAndGetFilters(r.URL.RawQuery)
	if err != nil {
		rh.log.Infof("Fail to parse given query : %v", err)
		return err
	}

	// apply filters to fetched repos
	filteredRepos := utils.ApplyFilters(githubRepos, filters...)

	err = json.NewEncoder(w).Encode(filteredRepos)
	if err != nil {
		rh.log.WithError(err).Error("Fail to encode JSON")
		return err
	}
	return nil
}

// fetchRepos call GitHup API for N latest created repos, then fetch language data for each, factorised for readability
func (rh *ReposHandler) fetchRepos() ([]*structs.GithubRepo, error) {
	repoSearchResults, err := rh.fetchNLatestCreatedRepos(rh.numRepos)
	if err != nil {
		rh.log.Infof("Fail to fetch Github repositories : %v", err)
		return nil, err
	}

	githubRepos := rh.fetchReposInfo(repoSearchResults, rh.numRepos)
	return githubRepos, nil
}

// fetchNLatestCreatedRepos fetch N latest created repos
func (rh *ReposHandler) fetchNLatestCreatedRepos(nRepos int) ([]*github.Repository, error) {
	repos := make([]*github.Repository, 0, nRepos)
	timeInterval := -10 * time.Minute
	endTime := time.Now().UTC()
	startTime := endTime.Add(timeInterval)

	for len(repos) < nRepos {
		query := fmt.Sprintf("created:%s..%s", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

		opts := &github.SearchOptions{
			Order: "desc",
			ListOptions: github.ListOptions{
				PerPage: nRepos,
			},
		}

		result, _, err := rh.client.Search.Repositories(context.Background(), "is:public "+query, opts)
		if err != nil {
			var rateLimitError *github.RateLimitError
			if errors.As(err, &rateLimitError) {
				rh.log.Fatal("Hit rate limit")
			}
			rh.log.Info("Fail to fetch Github repositories : %v", err)
			return nil, err
		}
		repos = append(repos, result.Repositories[:100-len(repos)]...)
		rh.log.Infof("Fetched %v repositories between %v and %v", len(repos), startTime, endTime)

		endTime = startTime
		startTime = endTime.Add(timeInterval)
	}
	return repos, nil
}

// fetchReposInfo concurrently retrieve language data of every repo
func (rh *ReposHandler) fetchReposInfo(searchResults []*github.Repository, numRepos int) []*structs.GithubRepo {
	githubRepos := make([]*structs.GithubRepo, 0, numRepos)
	// for each repo get info on each and launch go func and send to a channel
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, repo := range searchResults {
		wg.Add(1)
		go func(repository *github.Repository) {
			defer wg.Done()
			githubRepo := rh.getRepoInfo(repository)
			mu.Lock()
			defer mu.Unlock()
			githubRepos = append(githubRepos, githubRepo)
		}(repo)
	}
	wg.Wait()
	rh.log.Info("Fetched language information for repositories")
	return githubRepos
}

// getRepoInfo create GithubRepo struct containing all relevant info of an individual repo
func (rh *ReposHandler) getRepoInfo(repository *github.Repository) *structs.GithubRepo {
	license := strings.ToLower(repository.License.GetSPDXID()) // spdxid is a short descriptive format in kebab case child-commponnent childComponent

	gh := &structs.GithubRepo{
		Name:        repository.GetName(),
		Owner:       repository.GetOwner().GetLogin(),
		Url:         repository.GetURL(),
		Description: repository.GetDescription(),
		License:     license,
	}

	// fetch languages and corresponding bytes data
	languages, _, err := rh.client.Repositories.ListLanguages(context.Background(), repository.GetOwner().GetLogin(), repository.GetName())
	if err != nil {
		var rateLimitError *github.RateLimitError
		if errors.As(err, &rateLimitError) {
			rh.log.Fatal("Hit rate limit")
		}
		rh.log.Infof("Fail to fetch languages for owner %v's repository %v : %v", repository.GetOwner().GetLogin(), repository.GetName(), err)
		// return struct without languages if request fail
		return gh
	}

	languagesMap := make(map[string]int)
	for lan, bytes := range languages {
		languagesMap[strings.ToLower(lan)] = bytes
	}
	gh.SetLanguages(&languagesMap)

	return gh
}
