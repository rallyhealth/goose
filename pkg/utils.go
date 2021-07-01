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
	"fmt"
	"log"

	"github.com/AlecAivazis/survey/v2"
	"github.com/kireledan/gojenkins"
	color "github.com/logrusorgru/aurora/v3"
)

func getJenkinsJob(args []string) []string {
	if len(args) == 0 && IsGitDirectory() {
		fmt.Println(color.Yellow("Searching for job based on directory..."))
		return FindCurrentRepoJobs()
	} else {
		if len(args) == 0 {
			log.Fatal(color.Red("No job path specified. You must specify a job or run within a git repo."))
		}
		return []string{args[0]}
	}
}

func AutoLocateJob(args []string, branch string, Jenky *gojenkins.Jenkins) (*gojenkins.Job, string) {
	if len(args) < 1 {
		jobList := getJenkinsJob(args)
		chosenJob := ""
		chosenBranch := ""
		if len(jobList) > 1 {
			result := ""
			prompt := &survey.Select{
				Message: "Select jenkins job to run:",
				Help:    "Scroll or type the desired job",
				Options: jobList,
			}
			survey.AskOne(prompt, &result)
			chosenJob = result
		} else {
			if len(jobList) == 0 {
				log.Fatal(color.Red("No jobs found in jenkins for current repo/branch"))
			}
			chosenJob = jobList[0]
		}

		if chosenJob == "" {
			fmt.Println(color.Red("No job found or given"))
			return nil, ""
		}

		if branch != "" {
			chosenBranch = branch
		} else {
			chosenBranch = GetCurrentBranch()
		}

		jobPlusBranch := fmt.Sprintf("%s/%s", chosenJob, chosenBranch)
		thejob, _ := GetNestedJob(Jenky, jobPlusBranch)

		if thejob == nil {
			fmt.Println(color.Yellow("No job found for"), color.White(chosenBranch))
			fmt.Println(color.White("Locating PR branch..."))
			chosenBranch = fmt.Sprintf("PR-%s", GetCurrentPROfBranch())
			jobPlusBranch = fmt.Sprintf("%s/%s", chosenJob, chosenBranch)
			thejob, _ = GetNestedJob(Jenky, jobPlusBranch)
			if thejob != nil {
				fmt.Println(color.Green("Found job found for"), color.White(chosenBranch))
			}
		} else {
			fmt.Println(color.Green("Found job found for"), color.White(chosenBranch))
		}
		return thejob, jobPlusBranch
	} else {
		thejob, _ := GetNestedJob(Jenky, args[0])
		return thejob, args[0]
	}
}
