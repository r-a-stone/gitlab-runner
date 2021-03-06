---
last_updated: 2017-10-09
---

# Install GitLab Runner manually on GNU/Linux

If you can't use the [deb/rpm repository](linux-repository.md) to
install GitLab Runner, or your GNU/Linux OS is not among the supported
ones, you can install it manually using one of the methods below, as a
last resort.

If you want to use the [Docker executor](../executors/docker.md),
you must [install Docker](https://docs.docker.com/install/linux/docker-ce/centos/#install-docker-ce)
before using the Runner.

Make sure that you read the [FAQ](../faq/README.md) section which describes
some of the most common problems with GitLab Runner.

## Using deb/rpm package

It is possible to download and install via a `deb` or `rpm` package, if necessary.

### Download

To download the appropriate package for your system:

1. Find the latest file name and options at
   <https://gitlab-runner-downloads.s3.amazonaws.com/latest/index.html>.
1. Choose a version and download a binary, as described in the
   documentation for [downloading any other tagged
   releases](bleeding-edge.md#download-any-other-tagged-release) for
   bleeding edge GitLab Runner releases.

For example, for Debian or Ubuntu:

```sh
curl -LJO https://gitlab-runner-downloads.s3.amazonaws.com/latest/deb/gitlab-runner_<arch>.deb
```

For example, for CentOS or Red Hat Enterprise Linux:

```sh
curl -LJO https://gitlab-runner-downloads.s3.amazonaws.com/latest/rpm/gitlab-runner_<arch>.rpm
```

Note: **Note**
No arm64 deb/rpm packages are provided for GitLab Runner, but a [binary
file](#using-binary-file) is available. See the [related
issue](https://gitlab.com/gitlab-org/gitlab-runner/issues/4871) for more
information.

### Install

1. Install the package for your system as follows.

   For example, for Debian or Ubuntu:

   ```sh
   dpkg -i gitlab-runner_<arch>.deb
   ```

   For example, for CentOS or Red Hat Enterprise Linux:

   ```sh
   rpm -i gitlab-runner_<arch>.rpm
   ```

1. [Register the Runner](../register/index.md#gnulinux)

### Update

Download the latest package for your system then upgrade as follows:

For example, for Debian or Ubuntu:

```sh
dpkg -i gitlab-runner_<arch>.deb
```

For example, for CentOS or Red Hat Enterprise Linux:

```sh
rpm -Uvh gitlab-runner_<arch>.rpm
```

## Using binary file

It is possible to download and install via binary file, if necessary.

### Install

CAUTION: **Important:**
With GitLab Runner 10, the executable was renamed to `gitlab-runner`. If you
want to install a version prior to GitLab Runner 10, [visit the old docs](old.md).

1. Simply download one of the binaries for your system:

   ```sh
   # Linux x86-64
   sudo curl -L --output /usr/local/bin/gitlab-runner https://gitlab-runner-downloads.s3.amazonaws.com/latest/binaries/gitlab-runner-linux-amd64

   # Linux x86
   sudo curl -L --output /usr/local/bin/gitlab-runner https://gitlab-runner-downloads.s3.amazonaws.com/latest/binaries/gitlab-runner-linux-386

   # Linux arm
   sudo curl -L --output /usr/local/bin/gitlab-runner https://gitlab-runner-downloads.s3.amazonaws.com/latest/binaries/gitlab-runner-linux-arm

   # Linux arm64
   sudo curl -L --output /usr/local/bin/gitlab-runner https://gitlab-runner-downloads.s3.amazonaws.com/latest/binaries/gitlab-runner-linux-arm64
   ```

   You can download a binary for every available version as described in
   [Bleeding Edge - download any other tagged release](bleeding-edge.md#download-any-other-tagged-release).

1. Give it permissions to execute:

   ```sh
   sudo chmod +x /usr/local/bin/gitlab-runner
   ```

1. Create a GitLab CI user:

   ```sh
   sudo useradd --comment 'GitLab Runner' --create-home gitlab-runner --shell /bin/bash
   ```

1. Install and run as service:

   ```sh
   sudo gitlab-runner install --user=gitlab-runner --working-directory=/home/gitlab-runner
   sudo gitlab-runner start
   ```

1. [Register the Runner](../register/index.md)

NOTE: **Note**
If `gitlab-runner` is installed and run as service (what is described
in this page), it will run as root, but will execute jobs as user specified by
the `install` command. This means that some of the job functions like cache and
artifacts will need to execute `/usr/local/bin/gitlab-runner` command,
therefore the user under which jobs are run, needs to have access to the executable.

### Update

1. Stop the service (you need elevated command prompt as before):

   ```sh
   sudo gitlab-runner stop
   ```

1. Download the binary to replace Runner's executable. For example:

   ```sh
   sudo curl -L --output /usr/local/bin/gitlab-runner https://gitlab-runner-downloads.s3.amazonaws.com/latest/binaries/gitlab-runner-linux-amd64
   ```

   You can download a binary for every available version as described in
   [Bleeding Edge - download any other tagged release](bleeding-edge.md#download-any-other-tagged-release).

1. Give it permissions to execute:

   ```sh
   sudo chmod +x /usr/local/bin/gitlab-runner
   ```

1. Start the service:

   ```sh
   sudo gitlab-runner start
   ```
