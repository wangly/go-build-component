package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	baseSpace = "/go/src"
	cacheSpace = "/workflow-cache"
)

var (
	minConfidence = flag.Float64("min_confidence", 0.8, "minimum confidence of a problem to print it")
	setExitStatus = flag.Bool("set_exit_status", false, "set exit status to 1 if any issues are found")
	suggestions   int
)

type Builder struct {
	GitCloneURL string
	GitRef      string
	LintPackage string
	ProjectName string

	WorkflowCache bool

	workDir string
	gitDir  string
}

func NewBuilder(envs map[string]string) (*Builder, error) {
	b := &Builder{}
	if envs["GIT_CLONE_URL"] != "" {
		b.GitCloneURL = envs["GIT_CLONE_URL"]
		if b.GitRef = envs["GIT_REF"]; b.GitRef == "" {
			b.GitRef = "master"
		}
	} else if envs["_WORKFLOW_GIT_CLONE_URL"] != "" {
		b.GitCloneURL = envs["_WORKFLOW_GIT_CLONE_URL"]
		b.GitRef = envs["_WORKFLOW_GIT_REF"]
	} else {
		return nil, fmt.Errorf("environment variable GIT_CLONE_URL is required")
	}

	s := strings.TrimSuffix(strings.TrimSuffix(b.GitCloneURL, "/"), ".git")
	b.ProjectName = s[strings.LastIndex(s, "/")+1:]

	//fmt.Println(b.ProjectName)
	b.LintPackage = envs["LINT_PACKAGE"]

	b.WorkflowCache = strings.ToLower(envs["_WORKFLOW_FLAG_CACHE"]) == "true"

	if b.WorkflowCache {
		b.workDir = cacheSpace
	} else {
		b.workDir = baseSpace
	}
	b.gitDir = filepath.Join(b.workDir, b.ProjectName)

	return b, nil
}

func (b *Builder) run() error {
	if err := os.Chdir(b.workDir); err != nil {
		return fmt.Errorf("chdir to workdir (%s) failed:%v", b.workDir, err)
	}

	if _, err := os.Stat(b.gitDir); os.IsNotExist(err) {
		if err := b.gitPull(); err != nil {
			return err
		}

		if err := b.gitReset(); err != nil {
			return err
		}
	}

	if err := b.build(); err != nil {
		return err
	}
	return nil
}

func (b *Builder) gitPull() error {
	var command = []string{"git", "clone", "--recurse-submodules", b.GitCloneURL, b.ProjectName}
	if _, err := (CMD{Command: command}).Run(); err != nil {
		fmt.Println("Clone project failed:", err)
		return err
	}
	fmt.Println("Clone project", b.GitCloneURL, "succeed.")
	return nil
}

func (b *Builder) gitReset() error {
	cwd, _ := os.Getwd()
	fmt.Println("current: ", cwd)
	var command = []string{"git", "checkout", b.GitRef, "--"}
	if _, err := (CMD{command, b.gitDir}).Run(); err != nil {
		fmt.Println("Switch to commit", b.GitRef, "failed:", err)
		return err
	}
	fmt.Println("Switch to", b.GitRef, "succeed.")

	return nil
}

func (b *Builder) build() error {
	//cwd, _ := os.Getwd()
	var script string
	if b.LintPackage == "" {
		script = "gobuild `go build ./...`"
	} else {
		script = "gobuild " + b.LintPackage
	}
	var command = []string{"sh", "-c", script}
	if _, err := (CMD{command, b.gitDir}).Run(); err != nil {
		fmt.Printf("Exec: %s failed: %v", script, err)
	}
	fmt.Printf("Exec: %s succeed.\n", script)
	return nil
}

type CMD struct {
	Command []string // cmd with args
	WorkDir string
}

func (c CMD) Run() (string, error) {
	fmt.Println("Run CMD: ", strings.Join(c.Command, " "))

	cmd := exec.Command(c.Command[0], c.Command[1:]...)

	if c.WorkDir != "" {
		cmd.Dir = c.WorkDir
	}

	data, err := cmd.CombinedOutput()
	result := string(data)
	if len(result) > 0 {
		fmt.Println(result)
	}

	return result, err
}