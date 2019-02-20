package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/alecthomas/kingpin"
)

var (
	version   string
	branch    string
	revision  string
	buildDate string
	goVersion = runtime.Version()
)

var (
	// flags
	gitSource    = kingpin.Flag("git-source", "The source of the repository.").Envar("ESTAFETTE_GIT_SOURCE").Required().String()
	gitOwner     = kingpin.Flag("git-owner", "The owner of the repository.").Envar("ESTAFETTE_GIT_OWNER").Required().String()
	gitName      = kingpin.Flag("git-name", "The owner plus repository name.").Envar("ESTAFETTE_GIT_NAME").Required().String()
	gitBranch    = kingpin.Flag("git-branch", "The branch to clone.").Envar("ESTAFETTE_GIT_BRANCH").Required().String()
	buildVersion = kingpin.Flag("build-version", "The version currently building/releasing.").Envar("ESTAFETTE_BUILD_VERSION").Required().String()

	repoName   = kingpin.Flag("repo", "Set other repository name to clone from same owner.").Envar("ESTAFETTE_EXTENSION_REPO").Required().String()
	repoBranch = kingpin.Flag("branch", "Set other repository branch to clone from same owner.").Envar("ESTAFETTE_EXTENSION_BRANCH").String()

	bitbucketAPITokenJSON = kingpin.Flag("bitbucket-api-token", "Bitbucket api token credentials configured at the CI server, passed in to this trusted extension.").Envar("ESTAFETTE_CREDENTIALS_BITBUCKET_API_TOKEN").String()
	githubAPITokenJSON    = kingpin.Flag("github-api-token", "Github api token credentials configured at the CI server, passed in to this trusted extension.").Envar("ESTAFETTE_CREDENTIALS_GITHUB_API_TOKEN").String()
)

func main() {

	// parse command line parameters
	kingpin.Parse()

	// log to stdout and hide timestamp
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	// log startup message
	log.Printf("Starting estafette-extension-git-trigger version %v...", version)

	// log all environment variables
	log.Printf("%v", os.Environ())

	// get api token from injected credentials
	bitbucketAPIToken := ""
	if *bitbucketAPITokenJSON != "" {
		log.Println("Unmarshalling injected bitbucket api token credentials")
		var credentials []APITokenCredentials
		err := json.Unmarshal([]byte(*bitbucketAPITokenJSON), &credentials)
		if err != nil {
			log.Fatal("Failed unmarshalling injected bitbucket api token credentials: ", err)
		}
		if len(credentials) == 0 {
			log.Fatal("No bitbucket api token credentials have been injected")
		}
		bitbucketAPIToken = credentials[0].AdditionalProperties.Token
	}

	githubAPIToken := ""
	if *githubAPITokenJSON != "" {
		log.Println("Unmarshalling injected github api token credentials")
		var credentials []APITokenCredentials
		err := json.Unmarshal([]byte(*githubAPITokenJSON), &credentials)
		if err != nil {
			log.Fatal("Failed unmarshalling injected github api token credentials: ", err)
		}
		if len(credentials) == 0 {
			log.Fatal("No github api token credentials have been injected")
		}
		githubAPIToken = credentials[0].AdditionalProperties.Token
	}

	if *repoBranch == "" {
		*repoBranch = "master"
	}

	overrideGitURL := fmt.Sprintf("https://%v/%v/%v", *gitSource, *gitOwner, *repoName)
	if bitbucketAPIToken != "" {
		overrideGitURL = fmt.Sprintf("https://x-token-auth:%v@%v/%v/%v", bitbucketAPIToken, *gitSource, *gitOwner, *repoName)
	}
	if githubAPIToken != "" {
		overrideGitURL = fmt.Sprintf("https://x-access-token:%v@%v/%v/%v", githubAPIToken, *gitSource, *gitOwner, *repoName)
	}

	// git clone the specified repository branch to the specific directory
	err := gitCloneOverride(*repoName, overrideGitURL, *repoBranch, *repoName, true, 50)
	if err != nil {
		log.Fatalf("Error cloning git repository %v to branch %v into subdir %v: %v", *repoName, *repoBranch, *repoName, err)
	}

	err = gitCommitEmpty(*gitSource, *gitOwner, *gitName, *gitBranch, *buildVersion, *repoName)
	if err != nil {
		log.Fatalf("Error committing trigger commit for repository %v: %v", *repoName, err)
	}

	err = gitPush(*repoBranch, *repoName)
	if err != nil {
		log.Fatalf("Error pushing repository %v branch %v to origin: %v", *repoName, *repoBranch, err)
	}
}
