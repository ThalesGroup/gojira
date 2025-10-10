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
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/viper"
	"github.com/andygrunwald/go-jira"
	"github.com/danieljoos/wincred"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	gojiraCredentialsName = "gojira"
)

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(loginTokenCmd)
	loginCmd.AddCommand(deleteCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Create wincred credential to authenticate to a dedicated Jira Server / Project via Basic Authentication.",
	Long:  `Create wincred credential to authenticate to a dedicated Jira Server / Project via Basic Authentication.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.Set("gojiraAuthTransport", "basic")
		viper.WriteConfig()
		loginToJira()
	},
}

var loginTokenCmd = &cobra.Command{
	Use:   "tokenlogin",
	Short: "Create wincred credential to authenticate to a dedicated Jira Server / Project via Bearer Token.",
	Long:  `Create wincred credential to authenticate to a dedicated Jira Server / Project via Bearer Token.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.Set("gojiraAuthTransport", "bearer")
		viper.WriteConfig()
		loginToJira()
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete wincred credential to authenticate to a dedicated Jira Server / Project.",
	Long:  `Delete wincred credential to authenticate to a dedicated Jira Server / Project.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Deleting '%s' credential from wincred ... \n", gojiraCredentialsName)
		cred, err := wincred.GetGenericCredential(gojiraCredentialsName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving credential: %v\n", err)
			return
		}
		if err := cred.Delete(); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting credential: %v\n", err)
			return
		}
		fmt.Println("Credential deleted successfully.")
	},
}

func readSecureInput(prompt string, sensitive bool) (string, error) {
	for {
		fmt.Print(prompt)
		var input string
		var err error

		if sensitive {
			// Read sensitive input (no echo) for passwords/tokens.
			byteInput, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Printf("\nError reading input: %v\n", err)
				continue
			}
			input = strings.TrimSpace(string(byteInput))
		} else {
			// Read non-sensitive input (echoed) for URL/username.
			reader := bufio.NewReader(os.Stdin)
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Printf("\nError reading input: %v\n", err)
				continue
			}
			input = strings.TrimSpace(input) // Handles \n, \r for Windows/Linux.
		}

		if len(input) == 0 {
			fmt.Println("\nInput cannot be empty. Please try again.")
			continue
		}

		if sensitive {
			fmt.Printf("\nInput captured successfully (length: %d characters).\n", len(input))
		}
		return input, nil
	}
}

func loginToJira() *jira.Client {
	var client *jira.Client
	var username string

	cred, err := wincred.GetGenericCredential(gojiraCredentialsName)
	if err == nil {
		if viper.GetString("gojiraAuthTransport") == "bearer" {
			fmt.Printf("Using stored token for Jira URL: %s\n", string(cred.Attributes[0].Value))
			client = createJiraClientToken(string(cred.Attributes[0].Value), string(cred.CredentialBlob))
			username = cred.UserName // Use stored username
		} else {
			fmt.Printf("Using stored credentials for Jira URL: %s, Username: %s\n", string(cred.Attributes[0].Value), cred.UserName)
			client = createJiraClient(string(cred.Attributes[0].Value), cred.UserName, string(cred.CredentialBlob))
			username = cred.UserName
		}
		if client != nil {
			// Validate authentication
			user, _, err := client.User.GetSelf()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Authenticated as: %s\n", user.Name)
			// Ensure username is valid for reporter
			if username != "" {
				_, _, err = client.User.Get(username)
				if err != nil {
					//fmt.Fprintf(os.Stderr, "Stored username '%s' is invalid: %v\n", username, err)
					username = "" // Prompt for new username
				}
			}
			if username == "" {
				username = user.Name
				cred.UserName = username
				if err := cred.Write(); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to update credentials: %v\n", err)
					os.Exit(1)
				}
				viper.Set("username", username)
				if err := viper.WriteConfig(); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to write config: %v\n", err)
					os.Exit(1)
				}
			}
			return client
		}
	}

	// Create new credential
	fmt.Printf("Creating new credential\n")
	jiraURL, err := readSecureInput("Jira URL: ", false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read Jira URL: %v\n", err)
		os.Exit(1)
	}

	// Prompt for username in both basic and bearer modes
	username, err = readSecureInput("Jira Username: ", false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read username: %v\n", err)
		os.Exit(1)
	}

	password, err := readSecureInput("Jira "+ternary(viper.GetString("gojiraAuthTransport") == "basic", "Password", "Personal Access Token")+": ", true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", ternary(viper.GetString("gojiraAuthTransport") == "basic", "password", "token"), err)
		os.Exit(1)
	}

	cred = wincred.NewGenericCredential(gojiraCredentialsName)
	cred.CredentialBlob = []byte(password)
	cred.UserName = username // Store username for both auth types

	if viper.GetString("gojiraAuthTransport") == "basic" {
		client = createJiraClient(jiraURL, username, string(cred.CredentialBlob))
	} else {
		client = createJiraClientToken(jiraURL, string(cred.CredentialBlob))
	}

	if client == nil {
		fmt.Fprintf(os.Stderr, "Failed to create Jira client\n")
		os.Exit(1)
	}

	// Validate authentication and username
	user, _, err := client.User.GetSelf()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Authenticated as: %s\n", user.Name)

	// Validate username
	_, _, err = client.User.Get(username)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "Invalid username '%s': %v\n", username, err)
		//fmt.Printf("Using authenticated user: %s\n", user.Name)
		username = user.Name
		cred.UserName = username
	}

	credAttributes := []wincred.CredentialAttribute{
		{
			Keyword: "jiraUrl",
			Value:   []byte(jiraURL),
		},
	}
	cred.Attributes = credAttributes
	cred.Persist = wincred.PersistEnterprise

	if err := cred.Write(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to store credentials: %v\n", err)
		os.Exit(1)
	}

	viper.Set("username", username)
	viper.Set("jira_url", jiraURL)
	if err := viper.WriteConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Jira account was successfully logged on!")
	return client
}

func createJiraClient(jiraURL, username, password string) *jira.Client {
	if len(jiraURL) == 0 {
		var err error
		jiraURL, err = readSecureInput("Jira URL: ", false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read Jira URL: %v\n", err)
			return nil
		}
	}
	if len(username) == 0 {
		var err error
		username, err = readSecureInput("Jira Username: ", false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read username: %v\n", err)
			return nil
		}
	}
	if len(password) == 0 {
		var err error
		password, err = readSecureInput("Jira Password: ", true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read password: %v\n", err)
			return nil
		}
	}
	fmt.Printf("Creating Jira client with username: %s\n", username)
	tp := jira.BasicAuthTransport{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}

	jiraClient, err := jira.NewClient(tp.Client(), strings.TrimSpace(jiraURL))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while creating Jira Client: %v\n", err)
		return nil
	}
	return jiraClient
}

func createJiraClientToken(jiraURL, password string) *jira.Client {
	if len(jiraURL) == 0 {
		var err error
		jiraURL, err = readSecureInput("Jira URL: ", false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read Jira URL: %v\n", err)
			return nil
		}
	}
	if len(password) == 0 {
		var err error
		password, err = readSecureInput("Jira Personal Access Token: ", true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read token: %v\n", err)
			return nil
		}
	}

	tp := jira.BearerAuthTransport{
		Token: strings.TrimSpace(password),
	}

	jiraClient, err := jira.NewClient(tp.Client(), strings.TrimSpace(jiraURL))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while creating Jira Client: %v\n", err)
		return nil
	}
	return jiraClient
}

func ternary(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}