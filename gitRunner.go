package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func gitCloneOverride(gitName, gitURL, gitBranch, subdir string, shallowClone bool, shallowCloneDepth int) (err error) {

	log.Printf("Cloning git repository %v to branch %v into subdir %v with shallow clone is %v and depth %v...", gitName, gitBranch, subdir, shallowClone, shallowCloneDepth)

	// git clone
	err = gitCloneWithRetry(gitName, gitURL, gitBranch, shallowClone, shallowCloneDepth, subdir, 3)
	if err != nil {
		return
	}

	log.Printf("Finished cloning git repository %v to branch %v into subdir %v with shallow clone is %v and depth %v", gitName, gitBranch, subdir, shallowClone, shallowCloneDepth)

	return
}

func gitCloneWithRetry(gitName, gitURL, gitBranch string, shallowClone bool, shallowCloneDepth int, subdir string, retries int) (err error) {

	attempt := 0

	for attempt == 0 || (err != nil && attempt < retries) {

		err = gitClone(gitName, gitURL, gitBranch, shallowClone, shallowCloneDepth, subdir)
		if err != nil {
			log.Printf("Attempt %v cloning git repository %v to branch %v failed: %v", attempt, gitName, gitBranch, err)
		}

		// wait with exponential backoff
		<-time.After(time.Duration(math.Pow(2, float64(attempt))) * time.Second)

		attempt++
	}

	return
}

func gitClone(gitName, gitURL, gitBranch string, shallowClone bool, shallowCloneDepth int, subdir string) (err error) {

	targetDirectory := getTargetDir(subdir)

	args := []string{"clone", fmt.Sprintf("--branch=%v", gitBranch), gitURL, targetDirectory}
	if shallowClone {
		args = []string{"clone", fmt.Sprintf("--depth=%v", shallowCloneDepth), fmt.Sprintf("--branch=%v", gitBranch), gitURL, targetDirectory}
	}
	gitCloneCommand := exec.Command("git", args...)
	gitCloneCommand.Stdout = os.Stdout
	gitCloneCommand.Stderr = os.Stderr
	err = gitCloneCommand.Run()
	if err != nil {
		return
	}
	return
}

func getTargetDir(subdir string) string {
	return filepath.Join("/estafette-work", subdir)
}

func gitCommitEmpty(repoSource, repoOwner, repoName, repoBranch, buildVersion, subdir string) (err error) {

	log.Printf("Creating empty commit for repository %v/%v/%v...", repoSource, repoOwner, repoName)

	message := fmt.Sprintf("Triggered by %v/%v/%v, branch %v, version %v", repoSource, repoOwner, repoName, repoBranch, buildVersion)
	workdir := getTargetDir(subdir)

	args := []string{"commit", "--allow-empty", fmt.Sprintf("--message=%v", message)}
	gitCommitCommand := exec.Command("git", args...)
	gitCommitCommand.Stdout = os.Stdout
	gitCommitCommand.Stderr = os.Stderr
	gitCommitCommand.Dir = workdir
	err = gitCommitCommand.Run()
	if err != nil {
		return
	}
	return
}

func gitPush(gitBranch, subdir string) (err error) {

	log.Printf("Pushing empty commit for subdir %v to origin...", subdir)

	workdir := getTargetDir(subdir)

	args := []string{"push", "origin", gitBranch}
	gitPushCommand := exec.Command("git", args...)
	gitPushCommand.Stdout = os.Stdout
	gitPushCommand.Stderr = os.Stderr
	gitPushCommand.Dir = workdir
	err = gitPushCommand.Run()
	if err != nil {
		return
	}
	return
}

func setUser(username, email, subdir string) (err error) {

	log.Printf("Setting user name %v and email %v for repository in subdir %v...", username, email, subdir)

	workdir := getTargetDir(subdir)

	args := []string{"config", "user.email", email}
	gitConfigCommand := exec.Command("git", args...)
	gitConfigCommand.Stdout = os.Stdout
	gitConfigCommand.Stderr = os.Stderr
	gitConfigCommand.Dir = workdir
	err = gitConfigCommand.Run()
	if err != nil {
		return
	}

	args = []string{"config", "user.name", username}
	gitConfigCommand = exec.Command("git", args...)
	gitConfigCommand.Stdout = os.Stdout
	gitConfigCommand.Stderr = os.Stderr
	gitConfigCommand.Dir = workdir
	err = gitConfigCommand.Run()
	if err != nil {
		return
	}

	return
}
