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
	"fmt"

	"github.com/rallyhealth/goose/pkg"
	. "github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var rebuildIndex bool

// jobsCmd represents the jobs command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "search for a job associated with a url",
	Long: `search will print out the affiliated jenkins jobs with a git url.

	The format is github.com/[your argument].git

	Putting no argument will search your current working directory`,
	Run: func(cmd *cobra.Command, args []string) {

		if rebuildIndex == true {
			fmt.Println("Rebuilding job index...")
			pkg.BuildJobIndex(Jenky)
		}

		if len(args) == 0 {
			fmt.Println("Jobs Associated with ", Cyan("current directory"))
			fmt.Println(pkg.FindCurrentRepoJobs())
		}
		if len(args) == 1 {
			repo := fmt.Sprintf("github.com/%s.git", args[1])
			fmt.Println("Jobs Associated with : ", Cyan(repo))
			fmt.Println(pkg.GetAffiliatedJobs(repo))
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().BoolVarP(&rebuildIndex, "r", "r", false, "rebuild jobindex")
}
