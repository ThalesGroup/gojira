# GoJira
 [![Build Status](https://travis-ci.org/jurocknsail/gojira.svg?branch=master)](https://travis-ci.com/jurocknsail/gojira)

![gojira](https://github.com/jurocknsail/gojira/blob/master/logo/gojira-logo_150px.png)

## What is GoJira ?

GoJira is a Go tool designed to streamline and enhance the use of Jira, a popular project management software.
This version of GoJira focuses on enabling Definition of Done (DoD) injection.

## How to install GoJira ?

#### Download binary

* Visit the [Releases page](https://github.com/jurocknsail/gojira/releases/latest) to access the latest GoJira release.
* Copy the executable file to a directory in your system's PATH environment variable.
* For ease of use, the executable file should be renamed to "gojira.exe".

#### Verify the Executable

Verify the installation by opening your preferred command-line interface (CLI) and running the command "gojira". You should see output similar to the following:

```shell
$ gojira.exe
Gojira is a modern CLI Tool for Jira Server, allowing DoD injection and plenty other stuff impossible to achieve using Jira plugins.

Usage:
  gojira [flags]
  gojira [command]

Available Commands:
  config      Manage gojira configuration.
  dod         Apply Definition of Done to a set of US.
  help        Help about any command
  login       Create wincred credential to authenticate to a dedicated Jira Server / Project via Basic Authentication.
  pi-wl       List all PI sprints workload
  sprints     List all sprints in configured project.
  stories     List all User Stories in the configured Sprint.
  test
  tokenlogin  Create wincred credential to authenticate to a dedicated Jira Server / Project via Bearer Token.
  version     Print the version of Gojira

Flags:
  -h, --help   help for gojira

Use "gojira [command] --help" for more information about a command.

```
If you see the expected above output, the installation is complete.
Otherwise, ensure that you placed the gojira.exe file in a directory included in your system's PATH environment variable and added it correctly.

#### Configure gojira

* Set up Jira credentials: Configure your Jira credentials for authentication using gojira.exe login or gojira.exe tokenlogin.

```shell
$ gojira.exe login
Jira URL: https://jira.url.com
Jira Username: username
Jira Password: *********
Jira account was successfully logged on!
```
```shell
$ gojira.exe tokenlogin
Creating new credential
Jira URL: https://jira.url.com
Jira Username: bob
Jira Personal Access Token:
Input captured successfully (length: XX characters).
Authenticated as: bob
Jira account was successfully logged on!
```

* Configure the Jira Project: Specify the Jira project key that you want to work with using GoJira.

```shell
$ gojira.exe config project "<MY_PROJECT_NAME>"
Using stored token for Jira URL: https://jira.url.com
Authenticated as: bob
Several boards matching, please pick one by ID.
Board : {XXXXX https://jira.url.com/rest/agile/1.0/board/XXXXX MY_PROJECT_NAME scrum 0}
Board : {YYYYY https://jira.url.com/rest/agile/1.0/board/YYYYY OTHER_PROJECT_NAME scrum 0}
Selected Board ID: XXXXX
Board selected!
```

* Configure the Jira Project ID: Optionally, specify the numeric project ID (if known).

```shell
$ gojira.exe config projectId <MY_PROJECT_ID>
Using stored token for Jira URL: https://jira.url.com
Authenticated as: bob
Board selected!
Board ID: XXXXX
Board Name: MY_PROJECT_NAME
```

* Configure the sprint: Select the active sprint for the project. You can update this later by running the same command with a different sprint ID.

```shell
$ gojira.exe config sprint
Using stored token for Jira URL: https://jira.url.com
Authenticated as: bob
 Sprint ID : AAAAA : Sprint 1
 ...
 Sprint ID : CCCCC : Sprint 3
 ...
Please input desired sprint ID: CCCCC
```

#### Functionalities

##### List User Stories

* List all User Stories (US) in the selected sprint to identify those you want to modify with other gojira command.

```shell
$ gojira.exe stories
Using stored token for Jira URL: https://jira.url.com
Authenticated as: bob
Stories/Bugs in selected sprint :
 PROJ-DDDDD | Story  | A User Story
 PROJ-EEEEE | Story  | Another US   
 PROJ-FFFFF | Story  | Again a US   
 
As list :
PROJ-DDDDD,PROJ-EEEEE,PROJ-FFFFF
Total SP in sprint : XX
```

##### Apply Definition of Done (DoD)

* Apply a specified Definition of Done template to a list of User Stories.
* Usage : gojira dod [ dodname ] US-XXXXX,US-YYYYY,....

```shell
$ gojira.exe dod standardstory PROJ-DDDDD,PROJ-EEEEE
 Pushing standardstory DoD for US PROJ-DDDDD ...............
 Pushing standardstory DoD for US PROJ-EEEEE ...............
Number of US treated : 2
```

* Apply a specified Definition of Done template to an specific issue.
* Usage : gojira dodIssue [ dodname ] PROJ-XXXXX

```shell
$ gojira.exe dod standardstory PROJ-XXXXX
 Pushing standardstory DoD for issue PROJ-XXXXX ...............
Subtasks created for issue PROJ-XXXXX
```

## Additional Links

* [**Contributing**](./CONTRIBUTING.md)
* [**Development Guide**](./docs/developer-guide.md)

