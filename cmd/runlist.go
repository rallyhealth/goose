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
	"github.com/rallyhealth/goose/pkg"
	"github.com/spf13/cobra"
)

// runlistCmd represents the runlist command
var runlistCmd = &cobra.Command{
	Use:   "runlist",
	Short: "Run several jenkins jobs concurrently",
	Long: `
	Runlist will run jobs concurrently from a batchfile. For example:

"""
	cat <<EOF >>myBatchFile
BatchJobs:
  - Job: Path/To/My/Job
    Variables:
      VariableA: True
      VariableB: False
    Substitute:
      VariableC: [value1,value2]
EOF

goose runlist myBatchFile
"""

Will run 2 jobs concurrently for the 'Path/To/My/Job' on the main branch. The 'VariableA' and 'VariableB' variables will be the same for these 4 job runs. The end result is that each job is run with the same variables with 'VariableC' being different for each run.
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Run each command async OR sequentially
		jobList := pkg.ReadBatchJobFile(args[0])
		pkg.RunJobList(jobList, Jenky)
	},
}

func init() {
	rootCmd.AddCommand(runlistCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runlistCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runlistCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
