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
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
	"github.com/kireledan/gojenkins"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

func correctSlash(path string) string {
	newstring := strings.ReplaceAll(path, "//", "%2F")
	return newstring
}

func FollowBuild(B *gojenkins.Build) {
	re := NewRegexpWriter(os.Stdout)
	red := New(Brown, Black)
	re.AddRule(red, regexp.MustCompile(`\[Pipeline\]`))
	offset := int64(0)
	for B.IsRunning(context.TODO()) {
		time.Sleep(time.Second * 2)
		resp, _ := B.GetConsoleOutputFromIndex(context.TODO(), offset)
		offset = resp.Offset
		if len(resp.Content) > 0 {
			lines := strings.Split(resp.Content, "\n")
			for _, line := range lines {
				if !strings.Contains(line, "[Pipeline]") && !strings.Contains(line, "> git") {
					re.WriteString(line + "\n")
				}
			}
		}
	}
}

func GetJobParameters(J *gojenkins.Jenkins, path string) map[string]gojenkins.ParameterDefinition {
	nj, err := GetNestedJob(J, path)
	if err != nil {
		log.Fatal(err)
	}
	if nj != nil {
		params, _ := nj.GetParameters(context.TODO())
		paramlist := map[string]gojenkins.ParameterDefinition{}
		for _, param := range params {
			paramlist[param.Name] = param
		}
		return paramlist
	}
	fmt.Println(fmt.Sprintf("%s not found", path))
	return nil
}

func InvokeJob(Jenky *gojenkins.Jenkins, thejob *gojenkins.Job, choices map[string]string) {
	queueNum, _ := thejob.InvokeSimple(context.TODO(), choices)

	var build *gojenkins.Build
	spinner := wow.New(os.Stdout, spin.Get(spin.Dots), "Waiting to start build...")
	t, _ := Jenky.GetQueueItem(context.TODO(), queueNum)

	spinner.Start() // MAKE SURE TO STOP THIS!!
	for {
		if t.Raw.Executable.Number != 0 {
			spinner.PersistWith(spin.Spinner{Frames: []string{"✅   "}}, fmt.Sprintf("Started as build %d", t.Raw.Executable.Number))
		}
		runningBuild, _ := thejob.GetBuild(context.Background(), t.Raw.Executable.Number)
		if runningBuild != nil {
			spinner.PersistWith(spin.Spinner{Frames: []string{"✅"}}, fmt.Sprintf("Getting logs for %d", t.Raw.Executable.Number))
			build = runningBuild
			break
		}
		time.Sleep(2 * time.Second)
		t.Poll(context.TODO())
	}
	spinner.Stop()
	FollowBuild(build)
}

func GetNestedJob(J *gojenkins.Jenkins, path string) (*gojenkins.Job, error) {
	var currentJob *gojenkins.Job
	path = correctSlash(path)
	pathsplit := strings.Split(path, "/")
	changeTeamURL(J, pathsplit[0])
	for _, p := range pathsplit {
		fixedString := correctSlash(p)
		if currentJob == nil {
			currentJob, _ = J.GetJob(context.TODO(), fixedString)
		} else {
			currentJob, _ = currentJob.GetInnerJob(context.TODO(), fixedString)
		}
	}
	if currentJob == nil {
		return nil, errors.Errorf("no job found at %s", path)
	}
	return currentJob, nil
}

type Build struct {
	Job         string
	BuildNumber int64
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 32)
	return err == nil
}

func GetBuild(J *gojenkins.Jenkins, path string, num int64) *gojenkins.Build {
	j, err := GetNestedJob(J, path)
	if err != nil {
		log.Fatal(err)
	}
	testB, _ := j.GetBuild(context.TODO(), num)
	return testB
}

func ParseBuildPath(url string) Build {
	url = strings.Trim(url, "/")
	m := strings.Replace(url, "/job/", "/", -1)
	cleanedURL := strings.Split(m, "/")
	// The first part has the team name
	cleanedURL[0] = strings.Split(cleanedURL[0], "-")[1]

	// Then remove the last part
	buildnum, _ := strconv.ParseInt(cleanedURL[len(cleanedURL)-1], 10, 64)
	cleanedURL = cleanedURL[:len(cleanedURL)-1]

	// We dont need the beginning either
	cleanedURL = cleanedURL[1:]

	fullpath := ""
	for num, ele := range cleanedURL {
		fullpath += ele
		if num < len(cleanedURL)-1 {
			fullpath += "/"
		}
	}

	return Build{BuildNumber: buildnum, Job: fullpath}
}

func BuildScriptRequest(J *gojenkins.Jenkins, team string, script string) *http.Request {
	changeTeamURL(J, team)
	data := url.Values{}
	data.Set("script", script)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/scriptText", J.Server), bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.SetBasicAuth(J.Requester.BasicAuth.Username, J.Requester.BasicAuth.Password)

	return req
}

func RunScript(J *gojenkins.Jenkins, team string, script string) (string, error) {
	changeTeamURL(J, team)
	data := url.Values{}
	data.Set("script", script)

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/scriptText", J.Server), bytes.NewBufferString(data.Encode()))

	if err != nil {
		return "nil", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.SetBasicAuth(J.Requester.BasicAuth.Username, J.Requester.BasicAuth.Password)
	resp, err := client.Do(req)
	if err != nil {
		return "nil", err
	}

	text, _ := ioutil.ReadAll(resp.Body)

	return string(text), err
}

func GetInnerJobs(J *gojenkins.Jenkins, job *gojenkins.Job) []*gojenkins.Job {
	jobs, err := job.GetInnerJobs(context.TODO())
	if err != nil {
		return nil
	}
	return jobs
}

func changeTeamURL(J *gojenkins.Jenkins, team string) {
	rootJenkins := os.Getenv("JENKINS_ROOT_URL")
	J.Requester.Base = fmt.Sprintf("%s/teams-%s", rootJenkins, team)
	J.Server = fmt.Sprintf("%s/teams-%s", rootJenkins, team)
}

func GetJobs(J *gojenkins.Jenkins) []*gojenkins.Job {
	var joblist []*gojenkins.Job
	jobs, err := J.GetAllJobs(context.TODO())
	if err != nil {
		return nil
	}

	for _, job := range jobs {
		ljob := job
		joblist = append(joblist, ljob)
	}

	return joblist
}

func getLinks(body io.Reader) []string {
	var links []string
	z := html.NewTokenizer(body)
	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			//todo: links list shoudn't contain duplicates
			return links
		case html.StartTagToken, html.EndTagToken:
			token := z.Token()
			if "a" == token.Data {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
					}

				}
			}

		}
	}
}

type result struct {
	index int
	Res   http.Response
	err   error
	URL   string
}

// boundedParallelGet sends requests in parallel but only up to a certain
// limit, and furthermore it's only parallel up to the amount of CPUs but
// is always concurrent up to the concurrency limit
func BatchScript(requests []*http.Request, concurrencyLimit int) []result {

	// this buffered channel will block at the concurrency limit
	semaphoreChan := make(chan struct{}, concurrencyLimit)

	// this channel will not block and collect the http request results
	resultsChan := make(chan *result)

	// make sure we close these channels when we're done with them
	defer func() {
		close(semaphoreChan)
		close(resultsChan)
	}()
	client := &http.Client{}
	// keen an index and loop through every url we will send a request to
	for i, req := range requests {

		// start a go routine with the index and url in a closure
		rand.Seed(time.Now().UnixNano())
		r := rand.Intn(100)
		go func(i int, url *http.Request) {

			// this sends an empty struct into the semaphoreChan which
			// is basically saying add one to the limit, but when the
			// limit has been reached block until there is room
			semaphoreChan <- struct{}{}

			// send the request and put the response in a result struct
			// along with the index so we can sort them later along with
			// any error that might have occoured

			d := time.Duration(r) * time.Millisecond
			time.Sleep(d)
			res, err := client.Do(url)
			if *&res.StatusCode == 400 {
				fmt.Println("Failed... :(")
			}
			result := &result{i, *res, err, url.URL.RequestURI()}

			// now we can send the result struct through the resultsChan
			resultsChan <- result

			// once we're done it's we read from the semaphoreChan which
			// has the effect of removing one from the limit and allowing
			// another goroutine to start
			<-semaphoreChan

		}(i, req)
	}

	// make a slice to hold the results we're expecting
	var results []result

	// start listening for any results over the resultsChan
	// once we get a result append it to the result slice
	for {
		result := <-resultsChan
		results = append(results, *result)

		// if we've reached the expected amount of urls then stop
		if len(results) == len(requests) {
			break
		}
	}

	// let's sort these results real quick
	sort.Slice(results, func(i, j int) bool {
		return results[i].index < results[j].index
	})

	// now we're done we return the results
	return results
}
