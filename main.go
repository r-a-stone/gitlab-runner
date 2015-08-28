package main

import (
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"gitlab.com/gitlab-org/gitlab-ci-multi-runner/helpers/cli"
	"github.com/codegangsta/cli"

	"fmt"
	_ "gitlab.com/gitlab-org/gitlab-ci-multi-runner/commands"
	"gitlab.com/gitlab-org/gitlab-ci-multi-runner/common"
	_ "gitlab.com/gitlab-org/gitlab-ci-multi-runner/executors/docker"
	_ "gitlab.com/gitlab-org/gitlab-ci-multi-runner/executors/parallels"
	_ "gitlab.com/gitlab-org/gitlab-ci-multi-runner/executors/shell"
	_ "gitlab.com/gitlab-org/gitlab-ci-multi-runner/executors/ssh"
	_ "gitlab.com/gitlab-org/gitlab-ci-multi-runner/shells"
)

func main() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "a GitLab Runner"
	app.Version = fmt.Sprintf("%s (%s)", common.VERSION, common.REVISION)
	app.Author = "Kamil Trzciński"
	app.Email = "ayufan@ayufan.eu"
	cli_helpers.SetupLogLevelOptions(app)
	app.Commands = common.GetCommands()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
