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
	"io/ioutil"

	"github.com/kireledan/gojenkins"
	"github.com/rafaeldias/async"
	"gopkg.in/yaml.v2"
)

type BatchJob struct {
	Job        string              `yaml:"Job" `
	Variables  map[string]string   `yaml:"Variables" `
	Substitute map[string][]string `yaml:"Substitute"`
}

type JobList struct {
	Batch []BatchJob `yaml:"BatchJobs"`
}

func ReadBatchJob(content []byte) JobList {
	var config JobList
	err := yaml.Unmarshal(content, &config)
	check(err)
	ValidateBatchJob(config)
	return config
}

func ReadBatchJobFile(filename string) JobList {
	jobFileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return ReadBatchJob(jobFileContent)
}

func ValidateBatchJob(jl JobList) bool {
	for _, job := range jl.Batch {
		if !validateSubstituteLength(job) {
			return false
		}
	}
	return true
}

func validateSubstituteLength(job BatchJob) bool {
	substituteNumber := 0
	for _, substitutes := range job.Substitute {
		substituteNumber = len(substitutes)
		if len(substitutes) > 0 {
			if len(substitutes) != substituteNumber {
				fmt.Println("Job substitutes must be of equal length")
				return false
			}
		}
	}
	return true
}

func GenerateParameterList(jobBatch BatchJob) []map[string]string {
	// We iterate on each substitute
	batchLength := 0
	for _, substituteList := range jobBatch.Substitute {
		batchLength = len(substituteList)
		break
	}
	jobParameters := make([]map[string]string, batchLength)
	for key, substitute := range jobBatch.Substitute {
		for index, substituteValue := range substitute {
			// initialize nil map
			if jobParameters[index] == nil {
				jobParameters[index] = make(map[string]string)
			}
			jobParameters[index][key] = substituteValue
		}
	}
	for batchIndex, _ := range jobParameters {
		for variable, value := range jobBatch.Variables {
			jobParameters[batchIndex][variable] = value
		}
	}

	return jobParameters
}

func RunJobList(j JobList, Jenky *gojenkins.Jenkins) error {
	for _, job := range j.Batch {
		err := InvokeBatchJob(Jenky, job)
		if err != nil {
			return err
		}
	}
	return nil
}

func InvokeBatchJob(Jenky *gojenkins.Jenkins, jobBatch BatchJob) error {
	params := GenerateParameterList(jobBatch)
	job, err := GetNestedJob(Jenky, jobBatch.Job)
	if err != nil {
		panic(err)
	}
	taskList := async.Tasks{}

	for _, param := range params {
		paramCopy := param
		taskList = append(taskList, func() int {
			InvokeJob(Jenky, job, paramCopy)
			return 0
		})
	}

	fmt.Println("running batch jobs..")
	_, e := async.Concurrent(taskList)

	if e != nil {
		return e
	}

	return nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
