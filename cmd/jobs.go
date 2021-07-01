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
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/rallyhealth/goose/pkg"
	"github.com/kireledan/gojenkins"
	"github.com/spf13/cobra"
)

// jobsCmd represents the jobs command
var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "List jobs in a jenkins directory",
	Long: `List out the jobs in a jenkins directory. By default this will be the root of jenkins and list out the
	team jenkins.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			job, err := pkg.GetNestedJob(Jenky, args[0])
			if err != nil {
				log.Fatal(err)
			}
			jobs := job.Raw.Jobs
			if len(jobs) == 0 {
				fmt.Println("There are no jobs within this folder.")
				return
			}
			for _, job := range jobs {
				fmt.Println(job.Name)
			}
		} else {
			jobs, err := Jenky.GetAllJobNames(context.TODO())
			if err != nil {
				gojenkins.Error.Println(err)
			}
			for _, job := range jobs {
				fmt.Println(job.Name)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(jobsCmd)
}
