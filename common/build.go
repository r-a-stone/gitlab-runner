package common

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"gitlab.com/gitlab-org/gitlab-ci-multi-runner/helpers"
	"gitlab.com/hackwaw-disrupt/golang-proxy-server/Godeps/_workspace/src/golang.org/x/net/context"
)

type GitStrategy int

const (
	GitClone GitStrategy = iota
	GitFetch
	GitNone
)

type SubmoduleStrategy int

const (
	SubmoduleInvalid SubmoduleStrategy = iota
	SubmoduleNone
	SubmoduleNormal
	SubmoduleRecursive
)

type BuildRuntimeState string

const (
	BuildRunStatePending      BuildRuntimeState = "pending"
	BuildRunRuntimeRunning    BuildRuntimeState = "running"
	BuildRunRuntimeFinished   BuildRuntimeState = "finished"
	BuildRunRuntimeCanceled   BuildRuntimeState = "canceled"
	BuildRunRuntimeTerminated BuildRuntimeState = "terminated"
	BuildRunRuntimeTimedout   BuildRuntimeState = "timedout"
)

type BuildStage string

const (
	BuildStagePrepare           BuildStage = "prepare_script"
	BuildStageGetSources        BuildStage = "get_sources"
	BuildStageRestoreCache      BuildStage = "restore_cache"
	BuildStageDownloadArtifacts BuildStage = "download_artifacts"
	BuildStageUserScript        BuildStage = "build_script"
	BuildStageAfterScript       BuildStage = "after_script"
	BuildStageArchiveCache      BuildStage = "archive_cache"
	BuildStageUploadArtifacts   BuildStage = "upload_artifacts"
)

type Build struct {
	JobResponse `yaml:",inline"`

	Trace           JobTrace
	SystemInterrupt chan os.Signal `json:"-" yaml:"-"`
	RootDir         string         `json:"-" yaml:"-"`
	BuildDir        string         `json:"-" yaml:"-"`
	CacheDir        string         `json:"-" yaml:"-"`
	Hostname        string         `json:"-" yaml:"-"`
	Runner          *RunnerConfig  `json:"runner"`
	ExecutorData    ExecutorData

	// Unique ID for all running builds on this runner
	RunnerID int `json:"runner_id"`

	// Unique ID for all running builds on this runner and this project
	ProjectRunnerID int `json:"project_runner_id"`

	CurrentStage BuildStage
	CurrentState BuildRuntimeState

	executorStageResolver func() ExecutorStage
}

func (b *Build) Log() *logrus.Entry {
	return b.Runner.Log().WithField("job", b.ID).WithField("project", b.JobInfo.ProjectID)
}

func (b *Build) ProjectUniqueName() string {
	return fmt.Sprintf("runner-%s-project-%d-concurrent-%d",
		b.Runner.ShortDescription(), b.JobInfo.ProjectID, b.ProjectRunnerID)
}

func (b *Build) ProjectSlug() (string, error) {
	url, err := url.Parse(b.GitInfo.RepoURL)
	if err != nil {
		return "", err
	}
	if url.Host == "" {
		return "", errors.New("only URI reference supported")
	}

	slug := url.Path
	slug = strings.TrimSuffix(slug, ".git")
	slug = path.Clean(slug)
	if slug == "." {
		return "", errors.New("invalid path")
	}
	if strings.Contains(slug, "..") {
		return "", errors.New("it doesn't look like a valid path")
	}
	return slug, nil
}

func (b *Build) ProjectUniqueDir(sharedDir bool) string {
	dir, err := b.ProjectSlug()
	if err != nil {
		dir = fmt.Sprintf("project-%d", b.JobInfo.ProjectID)
	}

	// for shared dirs path is constructed like this:
	// <some-path>/runner-short-id/concurrent-id/group-name/project-name/
	// ex.<some-path>/01234567/0/group/repo/
	if sharedDir {
		dir = path.Join(
			fmt.Sprintf("%s", b.Runner.ShortDescription()),
			fmt.Sprintf("%d", b.ProjectRunnerID),
			dir,
		)
	}
	return dir
}

func (b *Build) FullProjectDir() string {
	return helpers.ToSlash(b.BuildDir)
}

func (b *Build) StartBuild(rootDir, cacheDir string, sharedDir bool) {
	b.RootDir = rootDir
	b.BuildDir = path.Join(rootDir, b.ProjectUniqueDir(sharedDir))
	b.CacheDir = path.Join(cacheDir, b.ProjectUniqueDir(false))
}

func (b *Build) executeStage(ctx context.Context, buildStage BuildStage, executor Executor) error {
	b.CurrentStage = buildStage

	shell := executor.Shell()
	if shell == nil {
		return errors.New("No shell defined")
	}

	script, err := GenerateShellScript(buildStage, *shell)
	if err != nil {
		return err
	}

	// Nothing to execute
	if script == "" {
		return nil
	}

	cmd := ExecutorCommand{
		Context: ctx,
		Script: script,
	}

	switch buildStage {
	case BuildStageUserScript, BuildStageAfterScript: // use custom build environment
		cmd.Predefined = false
	default: // all other stages use a predefined build environment
		cmd.Predefined = true
	}

	return executor.Run(cmd)
}

func (b *Build) executeUploadArtifacts(ctx context.Context, state error, executor Executor) (err error) {
	jobState := state

	for _, artifacts := range b.Artifacts {
		when := artifacts.When
		if state == nil {
			// Previous stages were successful
			if when == "" || when == ArtifactWhenOnSuccess || when == ArtifactWhenAlways {
				state = b.executeStage(ctx, BuildStageUploadArtifacts, executor)
			}
		} else {
			// Previous stage did fail
			if when == ArtifactWhenOnFailure || when == ArtifactWhenAlways {
				err = b.executeStage(ctx, BuildStageUploadArtifacts, executor)
			}
		}
	}

	// Use job's error if set
	if jobState != nil {
		err = jobState
	}
	return
}

func (b *Build) executeScript(ctx context.Context, executor Executor) error {
	// Prepare stage
	err := b.executeStage(ctx, BuildStagePrepare, executor)

	if err == nil {
		err = b.attemptExecuteStage(ctx, BuildStageGetSources, executor, b.GetGetSourcesAttempts())
	}
	if err == nil {
		err = b.attemptExecuteStage(ctx, BuildStageDownloadArtifacts, executor, b.GetDownloadArtifactsAttempts())
	}
	if err == nil {
		err = b.attemptExecuteStage(ctx, BuildStageRestoreCache, executor, b.GetRestoreCacheAttempts())
	}

	if err == nil {
		// Execute user build script (before_script + script)
		err = b.executeStage(ctx, BuildStageUserScript, executor)

		// Execute after script (after_script)
		timeoutContext, timeoutCancel := context.WithTimeout(ctx, AfterScriptTimeout)
		defer timeoutCancel()

		b.executeStage(timeoutContext, BuildStageAfterScript, executor)
	}

	// Execute post script (cache store, artifacts upload)
	if err == nil {
		err = b.executeStage(ctx, BuildStageArchiveCache, executor)
	}
	err = b.executeUploadArtifacts(ctx, err, executor)
	return err
}

func (b *Build) attemptExecuteStage(ctx context.Context, buildStage BuildStage, executor Executor, attempts int) (err error) {
	if attempts < 1 || attempts > 10 {
		return fmt.Errorf("Number of attempts out of the range [1, 10] for stage: %s", buildStage)
	}
	for attempt := 0; attempt < attempts; attempt++ {
		if err = b.executeStage(ctx, buildStage, executor); err == nil {
			return
		}
	}
	return
}

func (b *Build) GetBuildTimeout() time.Duration {
	buildTimeout := b.RunnerInfo.Timeout
	if buildTimeout <= 0 {
		buildTimeout = DefaultTimeout
	}
	return time.Duration(buildTimeout) * time.Second
}

func (b *Build) run(ctx context.Context, executor Executor) (err error) {
	b.CurrentState = BuildRunRuntimeRunning

	buildFinish := make(chan error, 1)

	runContext, runCancel := context.WithCancel(context.Background())
	defer runCancel()

	// Run build script
	go func() {
		buildFinish <- b.executeScript(runContext, executor)
	}()

	// Wait for signals: cancel, timeout, abort or finish
	b.Log().Debugln("Waiting for signals...")
	select {
	case <-b.Trace.Aborted():
		err = &BuildError{Inner: errors.New("canceled")}
		b.CurrentState = BuildRunRuntimeCanceled

	case <-ctx.Done():
		err = &BuildError{Inner: fmt.Errorf("execution took longer than %v seconds", b.GetBuildTimeout())}
		b.CurrentStage = BuildRunRuntimeTimedout

	case signal := <-b.SystemInterrupt:
		err = fmt.Errorf("aborted: %v", signal)
		b.CurrentState = BuildRunRuntimeTerminated

	case err = <-buildFinish:
		b.CurrentState = BuildRunRuntimeFinished
		return err
	}

	b.Log().WithError(err).Debugln("Waiting for build to finish...")

	// Wait till we receive that build did finish
	runCancel()
	<-buildFinish
	return err
}

func (b *Build) retryCreateExecutor(options ExecutorPrepareOptions, provider ExecutorProvider, logger BuildLogger) (executor Executor, err error) {
	for tries := 0; tries < PreparationRetries; tries++ {
		executor = provider.Create()
		if executor == nil {
			err = errors.New("failed to create executor")
			return
		}

		b.executorStageResolver = executor.GetCurrentStage

		err = executor.Prepare(options)
		if err == nil {
			break
		}
		if executor != nil {
			executor.Cleanup()
			executor = nil
		}
		if _, ok := err.(*BuildError); ok {
			break
		}

		logger.SoftErrorln("Preparation failed:", err)
		logger.Infoln("Will be retried in", PreparationRetryInterval, "...")
		time.Sleep(PreparationRetryInterval)
	}
	return
}

func (b *Build) CurrentExecutorStage() ExecutorStage {
	if b.executorStageResolver == nil {
		b.executorStageResolver = func() ExecutorStage {
			return ExecutorStage("")
		}
	}

	return b.executorStageResolver()
}

func (b *Build) Run(globalConfig *Config, trace JobTrace) (err error) {
	var executor Executor

	logger := NewBuildLogger(trace, b.Log())
	logger.Println(fmt.Sprintf("Running with %s\n  on %s (%s)", AppVersion.Line(), b.Runner.Name, b.Runner.ShortDescription()))

	b.CurrentState = BuildRunStatePending

	defer func() {
		if _, ok := err.(*BuildError); ok {
			logger.SoftErrorln("Job failed:", err)
			trace.Fail(err)
		} else if err != nil {
			logger.Errorln("Job failed (system failure):", err)
			trace.Fail(err)
		} else {
			logger.Infoln("Job succeeded")
			trace.Success()
		}
		if executor != nil {
			executor.Cleanup()
		}
	}()

	context, cancel := context.WithTimeout(context.Background(), b.GetBuildTimeout())
	defer cancel()

	options := ExecutorPrepareOptions{
		Config: b.Runner,
		Build: b,
		Trace: b.Trace,
		User: globalConfig.User,
		Context: context,
	}

	b.Trace = trace

	provider := GetExecutor(b.Runner.Executor)
	if provider == nil {
		return errors.New("executor not found")
	}

	executor, err = b.retryCreateExecutor(options, provider, logger)
	if err == nil {
		err = b.run(context, executor)
	}
	if executor != nil {
		executor.Finish(err)
	}
	return err
}

func (b *Build) String() string {
	return helpers.ToYAML(b)
}

func (b *Build) GetDefaultVariables() JobVariables {
	return JobVariables{
		{Key: "CI_PROJECT_DIR", Value: b.FullProjectDir(), Public: true, Internal: true, File: false},
		{Key: "CI_SERVER", Value: "yes", Public: true, Internal: true, File: false},
	}
}

func (b *Build) GetAllVariables() (variables JobVariables) {
	if b.Runner != nil {
		variables = append(variables, b.Runner.GetVariables()...)
	}
	variables = append(variables, b.GetDefaultVariables()...)
	variables = append(variables, b.Variables...)
	return variables.Expand()
}

func (b *Build) GetGitDepth() string {
	return b.GetAllVariables().Get("GIT_DEPTH")
}

func (b *Build) GetGitStrategy() GitStrategy {
	switch b.GetAllVariables().Get("GIT_STRATEGY") {
	case "clone":
		return GitClone

	case "fetch":
		return GitFetch

	case "none":
		return GitNone

	default:
		if b.AllowGitFetch {
			return GitFetch
		}

		return GitClone
	}
}

func (b *Build) GetSubmoduleStrategy() SubmoduleStrategy {
	if b.GetGitStrategy() == GitNone {
		return SubmoduleNone
	}
	switch b.GetAllVariables().Get("GIT_SUBMODULE_STRATEGY") {
	case "normal":
		return SubmoduleNormal

	case "recursive":
		return SubmoduleRecursive

	case "none", "":
		// Default (legacy) behavior is to not update/init submodules
		return SubmoduleNone

	default:
		// Will cause an error in AbstractShell) writeSubmoduleUpdateCmds
		return SubmoduleInvalid
	}
}

func (b *Build) IsDebugTraceEnabled() bool {
	trace, err := strconv.ParseBool(b.GetAllVariables().Get("CI_DEBUG_TRACE"))
	if err != nil {
		return false
	}

	return trace
}

func (b *Build) GetDockerAuthConfig() string {
	return b.GetAllVariables().Get("DOCKER_AUTH_CONFIG")
}

func (b *Build) GetGetSourcesAttempts() int {
	retries, err := strconv.Atoi(b.GetAllVariables().Get("GET_SOURCES_ATTEMPTS"))
	if err != nil {
		return DefaultGetSourcesAttempts
	}
	return retries
}

func (b *Build) GetDownloadArtifactsAttempts() int {
	retries, err := strconv.Atoi(b.GetAllVariables().Get("ARTIFACT_DOWNLOAD_ATTEMPTS"))
	if err != nil {
		return DefaultArtifactDownloadAttempts
	}
	return retries
}

func (b *Build) GetRestoreCacheAttempts() int {
	retries, err := strconv.Atoi(b.GetAllVariables().Get("RESTORE_CACHE_ATTEMPTS"))
	if err != nil {
		return DefaultRestoreCacheAttempts
	}
	return retries
}
