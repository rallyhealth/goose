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
	"net/http"
	"os"
	"time"

	"github.com/rallyhealth/goose/pkg"
	"github.com/kireledan/gojenkins"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var Jenky *gojenkins.Jenkins

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goose",
	Short: "A Jenkins CLI",
	Long: `The unofficial way to interact with Jenkins
	                                   ___
                               ,-""   '.
                             ,'  _   e )'-._
                            /  ,' '-._<.===-'   HONK HONK
                           /  /
                          /  ;
              _          /   ;
 ('._    _.-"" ""--..__,'    |
 <_  '-""                     \
  <'-                          :
   (__   <__.                  ;
     '-.   '-.__.      _.'    /
        \      '-.__,-'    _,'
         '._    ,    /__,-'
            ""._\__,'< <____
                 | |  '----.'.
                 | |        \ '.
                 ; |___      \-''
                 \   --<
                  '.'.<
                    '-'`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error
		user := os.Getenv("JENKINS_EMAIL")
		key := os.Getenv("JENKINS_API_KEY")
		rootJenkins := os.Getenv("JENKINS_ROOT_URL")
		loginJenkins := os.Getenv("JENKINS_LOGIN_URL")
		if user == "" || key == "" || rootJenkins == "" || loginJenkins == "" {
			fmt.Println("Please define $JENKINS_EMAIL and $JENKINS_API_KEY and $JENKINS_ROOT_URL and $JENKINS_LOGIN_URL")
			os.Exit(1)
		}
		clientWithTimeout := http.Client{
			Timeout: 3 * time.Second,
		}
		Jenky, err = gojenkins.CreateJenkins(&clientWithTimeout, fmt.Sprintf("%s", loginJenkins), user, key).Init(context.TODO())
		if err != nil {
			fmt.Println("Error connecting to jenkins. Make sure you are able to reach your jenkins instance. Is it on a VPN?")
			log.Fatal(err)
		}
		status, err := Jenky.Poll(context.TODO())
		if status == 401 {
			log.Fatal("Invalid credentials. Double check your jenkins envs JENKINS_EMAIL and JENKINS_API_KEY")
		}
		pkg.RefreshJobIndex(Jenky)

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	viper.SetDefault("author", "kireledan erik.nadel@rallyhealth.com")

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.goose.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".goose" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".goose")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
