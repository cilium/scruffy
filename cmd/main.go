// Copyright 2021 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"time"

	"github.com/cilium/scruffy/pkg/quay"
	"golang.org/x/sync/semaphore"

	flag "github.com/spf13/pflag"
)

var (
	orgName, gitRepository string
	repositories           []string
	gitBranches            []string
	expiration             time.Duration
	allImages              = []string{
		"cilium-ci",
		"clustermesh-apiserver-ci",
		"docker-plugin-ci",
		"hubble-relay-ci",
		"kvstoremesh-ci",
		"operator-ci",
		"operator-generic-ci",
		"operator-azure-ci",
		"operator-alibabacloud-ci",
		"operator-aws-ci",
	}
	stableBranches = []string{
		"origin/master",
		"origin/v1.12",
		"origin/v1.11",
		"origin/v1.10",
		"origin/v1.9",
		"origin/v1.8",
		"origin/v1.7",
	}
)

func init() {
	flag.StringVar(&orgName, "organization", "cilium", "Quay organization name")
	flag.StringSliceVar(&repositories, "repositories", allImages, "Quay repositories names separated by a comma")
	flag.StringSliceVar(&gitBranches, "stable-branches", stableBranches, "Quay repositories names separated by a comma")
	flag.StringVar(&gitRepository, "git-repository", "", "Git repository location that contain the SHAs that should not be deleted")
	flag.DurationVar(&expiration, "expiration", 168*time.Hour /* 1 week */, "Mark for expiration, set 0 to disable")
	flag.Parse()

	go signals()
}

var globalCtx, cancel = context.WithCancel(context.Background())

func signals() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	<-signalCh
	cancel()
}

func main() {
	if gitRepository == "" {
		panic("Empty git repository")
	}
	gitCmd, err := exec.LookPath("git")
	if err != nil {
		panic("Unable to find git command")
	}
	sm := semaphore.NewWeighted(10)
	quayToken := os.Getenv("QUAY_TOKEN")
	tagRegex := regexp.MustCompile(`[0-9a-f]{40}(-race)?`)
	now := time.Now()
	for _, repository := range repositories {
		var page = 1
		for {
			u := fmt.Sprintf("https://quay.io/api/v1/repository/%s/%s/tag/?onlyActiveTags=true&page=%d", orgName, repository, page)
			select {
			case <-globalCtx.Done():
				return
			default:
			}
			res, err := http.Get(u)
			if err != nil {
				panic(err)
			}
			var tg quay.TagsGet
			err = json.NewDecoder(res.Body).Decode(&tg)
			if err != nil {
				panic(err)
			}
			res.Body.Close()
			for _, tag := range tg.Tags {
				if !tagRegex.MatchString(tag.Name) {
					continue
				}
				var found bool
				for _, stableBranch := range gitBranches {
					cmd := exec.Command(gitCmd, "merge-base", "--is-ancestor", tag.Name, stableBranch)
					cmd.Dir = gitRepository
					err := cmd.Run()
					found = err == nil
					if found {
						fmt.Printf("Commit SHA found in git repo on branch %q, keeping tag %q\n", stableBranch, tag.Name)
						break
					}
				}
				if !found && tag.Expiration.IsZero() {
					if expiration != 0 {
						sm.Acquire(context.Background(), 1)
						go func(repositoryName, tagName string) {
							defer sm.Release(1)
							expire(now, expiration, repositoryName, tagName, quayToken)
						}(repository, tag.Name)
					} else {
						fmt.Printf("Tag %q set to expire: %t\n", tag.Name, !tag.Expiration.IsZero())
					}
				}
			}
			if !tg.HasAdditional {
				break
			}
			page++
		}
	}
}

func expire(now time.Time, expiration time.Duration, repository, tagName, quayToken string) {
	expire := now.Add(expiration)
	fmt.Printf("Marking tag %q for expiration at: %s\n", tagName, expire.Format(time.RFC3339))
	u := fmt.Sprintf("https://quay.io/api/v1/repository/%s/%s/tag/%s", orgName, repository, tagName)
	r := strings.NewReader(fmt.Sprintf(`{"expiration":%d}`, expire.Unix()))
	req, err := http.NewRequest(http.MethodPut, u, r)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+quayToken)
	req.Header.Add("Content-Type", "application/json")
	select {
	case <-globalCtx.Done():
		return
	default:
	}
	res, err := http.DefaultClient.Do(req)
	rb, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	res.Body.Close()
	fmt.Printf("Response for %q: %s", tagName, string(rb))
}
