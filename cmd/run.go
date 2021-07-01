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
	"strconv"
	"strings"

	tm "github.com/buger/goterm"

	"github.com/AlecAivazis/survey/v2"
	"github.com/jedib0t/go-pretty/table"
	. "github.com/logrusorgru/aurora/v3"
	"github.com/rallyhealth/goose/pkg"
	"github.com/spf13/cobra"
)

// runCmd represents the run command

var runCmd = &cobra.Command{
	Use:   "run",
	Args:  cobra.ArbitraryArgs,
	Short: "Run a jenkins job",
	Long: `Run a jenkins job based on the current git repo and branch.

	Otherwise, you can specify a job to be run.
	For example, Example/kireledan//long-branch-name specifies the Job , 'Example' , given the branch kireledan/long-branch-name
	Note that '/' in git branches are replaced  with '//'`,
	Run: func(cmd *cobra.Command, args []string) {
		thejob, jobPath := pkg.AutoLocateJob(args, cmd.Flag("branch").Value.String(), Jenky)

		params := pkg.GetJobParameters(Jenky, jobPath)
		choices := map[string]string{}

		interactive, _ := strconv.ParseBool(cmd.Flag("interactive").Value.String())
		if !interactive {
			jobArgs := map[string]*string{}
			for _, param := range params {
				tester := cmd.LocalFlags().String(param.Name, fmt.Sprintf("%v", param.DefaultParameterValue.Value), param.Description)
				jobArgs[param.Name] = tester
			}
			err := cmd.LocalFlags().Parse(args)
			if err != nil {
				cmd.LocalFlags().Args()
				panic(err)
			}
			for arg, argstuff := range jobArgs {
				fmt.Println(arg, *argstuff)
				choices[arg] = *argstuff
			}

		} else {
			// Pick params

			fmt.Println("Please pick your parameters for", thejob.Raw.FullDisplayName)
			for _, param := range params {
				if len(param.Choices) > 0 {
					result := ""
					prompt := &survey.Select{
						Message: fmt.Sprintf("%s: ", param.Name),
						Help:    param.Description,
						Options: param.Choices,
						Default: param.DefaultParameterValue.Value,
					}
					survey.AskOne(prompt, &result)
					choices[param.Name] = result
				} else {
					result := ""
					prompt := &survey.Input{
						Message: fmt.Sprintf("%s: ", param.Name),
						Help:    param.Description,
						Default: fmt.Sprintf("%v", param.DefaultParameterValue.Value),
					}
					survey.AskOne(prompt, &result)
					choices[param.Name] = result
				}

			}
			tm.Clear()
			tm.Flush()
			paramTable := table.NewWriter()
			//paramTable.SetOutputMirror(os.Stdout)
			paramTable.AppendHeader(table.Row{"PARAMETER", "CHOICE"})
			for name, choice := range choices {
				if choice == "" {
					choice = "None"
				}
				paramTable.AppendRow(table.Row{name, choice})
			}
			output := paramTable.Render()
			shortcut := printShortCut(jobPath, choices)

			fmt.Println(shortcut)
			fmt.Println("----")
			// Add some content to the box
			// Note that you can add ANY content, even tables
			job := strings.Split(thejob.Raw.FullName, "/")[:len(strings.Split(thejob.Raw.FullName, "/"))-1]
			branch := strings.Split(thejob.Raw.FullName, "/")[len(strings.Split(thejob.Raw.FullName, "/"))-1]
			fmt.Print("Job: ", Cyan(job), "\n")
			fmt.Print("Branch: ", Cyan(branch), "\n")
			fmt.Print(output)
			fmt.Print("\nPlease confirm these parameters by pressing ", Yellow("[Enter]"), "\n")

			// Move Box to approx center of the screen

			fmt.Scanln()
		}

		pkg.InvokeJob(Jenky, thejob, choices)
	},
}

func printShortCut(job string, choices map[string]string) string {
	builderString := fmt.Sprintf("Next time, type this to run this build without a prompt!\n")
	builderString += fmt.Sprintf("%s %s run %s --interactive=False -- ", Yellow("$"), Green("goose"), Cyan(job))

	for arg, value := range choices {
		if value != "" {
			builderString += fmt.Sprintf("--%s='%s' ", arg, value)
		}
	}

	return builderString
}

func init() {

	rootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolP("interactive", "i", true, "Help message for toggle")
	runCmd.Flags().StringP("branch", "b", "", "Help message for toggle")
}

func getJenkinsJob(args []string) []string {
	if len(args) == 0 {
		return pkg.FindCurrentRepoJobs()
	} else {
		return []string{args[0]}
	}
}
