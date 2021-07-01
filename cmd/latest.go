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
	"time"

	"github.com/rallyhealth/goose/pkg"
	"github.com/spf13/cobra"
)

var logs bool

// latestCmd represents the latest command
var latestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Grabs the latest build of a jenkins job",
	Long: `latest will print out the URL of the last running or currently running build of a given job.
	If the job is still running, the output will be streamed to the terminal. If not, the output from the build will be printed in the terminal.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, jobPath := pkg.AutoLocateJob(args, "", Jenky)

		jenkinsJob, err := pkg.GetNestedJob(Jenky, jobPath)
		if err != nil {
			log.Fatal(err)
		}
		b, err := jenkinsJob.GetLastBuild(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(b.GetUrl())
		if b.IsRunning(context.TODO()) {
			offset := int64(0)
			for b.IsRunning(context.TODO()) {
				time.Sleep(time.Second * 2)
				resp, _ := b.GetConsoleOutputFromIndex(context.TODO(), offset)
				offset = resp.Offset
				if len(resp.Content) > 0 {
					fmt.Println(resp.Content)
				}
			}
		} else {
			fmt.Println(b.GetConsoleOutput(context.TODO()))
		}
	},
}

func init() {
	rootCmd.AddCommand(latestCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// latestCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	latestCmd.Flags().BoolVarP(&logs, "logs", "l", false, "show logs")
}
