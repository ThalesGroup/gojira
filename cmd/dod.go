// Copyright 2019 Gemalto. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	helpers "github.com/jurocknsail/gojira/helpers"
)

func init() {
	rootCmd.AddCommand(dodCmd)
	dodCmd.AddCommand(listCmd)
	rootCmd.AddCommand(dodIssueCmd)
	dodIssueCmd.AddCommand(listCmd)
}

var dodCmd = &cobra.Command{
	Use:   "dod",
	Short: "Apply Definition of Done to a set of US.",
	Long:  `Apply Definition of Done to a set of US. The project and sprint used are the one configured using 'gojira config' commands.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("Requires DoD type and a list of Stories")
		}
		if isValidDoDType(args[0]) {
			return nil
		}
		return fmt.Errorf("Invalid dod type. Please use 'gojira dod list' to get all dods available.")
	},
	// ValidArgs: []string{"feature", "bug", "sprint", "pi", "study", "archi", "vlr"},
	Example: "gojira dod myDodName US-XXXXX,US-YYYYY,....",
	Run: func(cmd *cobra.Command, args []string) {

		sprintID := viper.GetString("sprint_id")
		username := viper.GetString("username")

		if sprintID != "" {

			jiraClient := loginToJira()
			issuesFound, _ := listStoriesForSprint(jiraClient, sprintID)

			usList := strings.Trim(strings.ToUpper(args[1]), " ")

			usInjected := 0
			for _, issue := range issuesFound {
				if strings.Contains(usList, issue.Key) {
					fmt.Printf(" Pushing %s DoD for US %s ", args[0], issue.Key)

					viper.SetConfigName("dod")
					viper.ReadInConfig()

					tasks := viper.GetStringSlice(args[0])

					//Restore config initial context
					viper.SetConfigName("gojira")
					viper.ReadInConfig()

					for _, summary := range tasks {
						helpers.CreateSubTask(jiraClient, issue.Fields.Project.Key, username, issue.Key, issue.ID, summary)
						fmt.Printf(".")
					}
					usInjected = usInjected + 1
					fmt.Printf("\n")
				}
			}
			fmt.Printf("Number of US treated : %d\n", usInjected)

		} else {
			fmt.Printf("Please configure a sprint first using 'gojira config sprint' command ! \n")
		}

	},
}

var dodIssueCmd = &cobra.Command{
    Use:   "dodIssue",
    Short: "Apply Definition of Done to a specific Jira ticket.",
    Long:  `Apply Definition of Done to a specific Jira ticket. The project is configured using 'gojira config' commands.`,
    Args: func(cmd *cobra.Command, args []string) error {
        if len(args) != 2 {
            return errors.New("Requires DoD type and a single Jira issue key (e.g., PROJ-XXXXX)")
        }
        if isValidDoDType(args[0]) {
            return nil
        }
        return fmt.Errorf("Invalid DoD type. Please use 'gojira dod list' to get all DoDs available.")
    },
    Example: "gojira dodIssueCmd myDodName PROJ-XXXXX",
    Run: func(cmd *cobra.Command, args []string) {
        username := viper.GetString("username")

        // Get the ticket key from arguments
        ticketKey := strings.Trim(strings.ToUpper(args[1]), " ")

        jiraClient := loginToJira()

        // Fetch the specific ticket
        issue, _, err := jiraClient.Issue.Get(ticketKey, nil)
        if err != nil {
            fmt.Printf("Error: Could not find issue %s. Please verify the issue key. Error: %v\n", ticketKey, err)
            return
        }

        fmt.Printf("Pushing %s DoD for issue %s\n", args[0], issue.Key)

        // Load DoD tasks from config
        viper.SetConfigName("dod")
        viper.ReadInConfig()

        tasks := viper.GetStringSlice(args[0])

        // Restore config initial context
        viper.SetConfigName("gojira")
        viper.ReadInConfig()

        // Create subtasks for the ticket
        for _, summary := range tasks {
            helpers.CreateSubTask(jiraClient, issue.Fields.Project.Key, username, issue.Key, issue.ID, summary)
            fmt.Printf(".")
        }
        fmt.Printf("\n")
        fmt.Printf("Subtasks created for issue %s\n", issue.Key)
    },
}

func isValidDoDType(dodType string) bool {

	viper.SetConfigName("dod")
	viper.ReadInConfig()
	
	allowedDoDTypes := viper.AllKeys()
	
	//Restore config initial context
	viper.SetConfigName("gojira")
	viper.ReadInConfig()
	
	if strings.Contains(strings.Join(allowedDoDTypes," "), dodType) {
		return true
	}
	return false
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available DoD in config.",
	Long:  `List available DoD in config.`,
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		viper.SetConfigName("dod")
		viper.ReadInConfig()

		fmt.Printf("\n")
		for _, dodType := range viper.AllKeys() {
			fmt.Printf(" %s \n", dodType)
			for _, task := range viper.GetStringSlice(dodType) {
				fmt.Printf("  - %s \n", task)
			}
			fmt.Printf("\n")
		}
		//Restore config initial context
		viper.SetConfigName("gojira")
		viper.ReadInConfig()
	},
}
