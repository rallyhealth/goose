/*
MIT License

Copyright (c) 2021 Rally Health, Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package pkg

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/kireledan/gojenkins"
	"github.com/prologic/bitcask"
)

const Query = `import org.jenkinsci.plugins.workflow.job.WorkflowJob;

def printScm(project, scm){
    if (scm instanceof hudson.plugins.git.GitSCM) {
        scm.getRepositories().each {
            it.getURIs().each {
                println(project + "\t"+ it.toString());
            }
        }
    }
}

Jenkins.instance.getAllItems(Job.class).each {
    project = it.getFullName()
   if (it instanceof WorkflowJob) {
        it.getSCMs().each {
            printScm(project, it)
      }
   }

}`

const datadelimiter = ","

type JobIndex struct {
	db *bitcask.Bitcask
}

func openDB() JobIndex {
	db, _ := bitcask.Open("/tmp/jenkins-db")
	return JobIndex{db}
}

func (j JobIndex) closeDB() {
	j.db.Close()
}

func RefreshJobIndex(j *gojenkins.Jenkins) {
	db := openDB()
	if db.db.Len() == 0 {
		db.closeDB()
		fmt.Println("Job index empty. Updating...")
		BuildJobIndex(j)
	} else {
		db.closeDB()
	}
}

func GetAffiliatedJobs(url string) []string {
	db := openDB()
	defer db.closeDB()
	return db.getAffiliatedJobs(url)
}

func FindCurrentRepoJobs() []string {
	return GetAffiliatedJobs(GetCurrentGitRepo())
}

func GetCurrentGitRepo() string { // repo, branch
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("")
	}
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return ""
	}
	remote, err := repo.Remote("origin")
	if err != nil {
		fmt.Println("Git repo remote is not set to origin! Unable to lookup remote.")
		return ""
	}
	remoteURL := strings.Replace(remote.Config().URLs[0], "git@", "", -1)
	remoteURL = strings.Replace(remoteURL, "github.com:", "github.com/", -1)
	return remoteURL
}

func GetCurrentPROfBranch() string {
	//git ls-remote origin 'pull/*/head' | grep -F -f <(git rev-parse HEAD) | awk -F'/' '{print $3}'
	dir, _ := os.Getwd()
	repo, _ := git.PlainOpen(dir)
	err := repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"+refs/pull/*/head:refs/remotes/origin/pr/*"},
	})

	if err != nil {
		if err.Error() != "already up-to-date" {
			fmt.Println(err)
		}
	}

	refs, err := repo.References()
	headref, _ := repo.Head()
	PRNumber := ""
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		// The HEAD is omitted in a `git show-ref` so we ignore the symbolic
		// references, the HEAD
		if ref.Type() == plumbing.SymbolicReference {
			return nil
		}
		if ref.Hash().String() == headref.Hash().String() && strings.Contains(ref.Name().String(), "/pr/") {
			branch := ref.Name().String()
			PRNumber = branch[strings.LastIndex(branch, "/")+1:]
		}

		return nil
	})

	return PRNumber
}

func IsGitDirectory() bool {
	dir, _ := os.Getwd()
	repo, _ := git.PlainOpen(dir)
	return repo != nil
}

func GetCurrentBranch() string {
	dir, _ := os.Getwd()
	repo, _ := git.PlainOpen(dir)
	tr, _ := repo.Head()
	branch := tr.Name().String()
	fixed := strings.Replace(branch, "refs/heads/", "", -1)
	fixed = strings.Replace(fixed, "%2F", "/", -1)
	return fixed
}

func (j JobIndex) addAffiliatedJob(url string, jenkinsJob string) {
	if !j.isAlreadyAssociated(url, jenkinsJob) {
		res, _ := j.db.Get([]byte(url))
		newList := string(res) + fmt.Sprintf("%s%s", datadelimiter, jenkinsJob)
		j.db.Put([]byte(url), []byte(newList))
	}
}

func (j JobIndex) getAffiliatedJobs(url string) []string {
	res, _ := j.db.Get([]byte(url))

	splitFn := func(c rune) bool {
		return c == ','
	}
	return strings.FieldsFunc(string(res), splitFn)
}

func (j JobIndex) isAlreadyAssociated(url string, jenkinsJob string) bool {
	associations, _ := j.db.Get([]byte(url))
	for _, association := range strings.Split(string(associations), datadelimiter) {
		if association == jenkinsJob {
			return true
		}
	}
	return false
}

func BuildJobIndex(Jenky *gojenkins.Jenkins) {
	//JobRepoMap := map[string][]string{}
	jobs, _ := Jenky.GetAllJobNames(context.TODO())
	requests := []*http.Request{}

	for _, job := range jobs {
		req := BuildScriptRequest(Jenky, job.Name, Query)
		requests = append(requests, req)
	}

	//fmt.Println(requests)

	jb := openDB()
	defer jb.closeDB()

	results := BatchScript(requests, 30)
	for _, res := range results {
		results, _ := ioutil.ReadAll(res.Res.Body)
		removedEndingEcho := strings.Split(string(results), "Result: ")
		cleanedUpList := strings.Split(removedEndingEcho[0], "\n")
		//teamURL := strings.Replace(res.URL, "/scriptText", "", -1)
		//team := strings.Replace(teamURL, "/teams-", "", -1)
		for _, item := range cleanedUpList {
			jobURLSplit := strings.Split(item, "\t")
			if len(jobURLSplit) == 2 {
				job := jobURLSplit[0]
				gitURL := strings.Replace(jobURLSplit[1], "https://", "", -1)
				branchIndex := strings.LastIndex(job, "/")
				if branchIndex > 0 {
					jobURLMinusBranch := job[0:branchIndex]
					//fullJobURL := fmt.Sprintf("%s", team, jobURLMinusBranch)
					//fmt.Println(gitURL)
					jb.addAffiliatedJob(gitURL, jobURLMinusBranch)
				}
			}
		}
	}
	fmt.Println("Job index completed.")
}
