package volumes

import (
	"crypto/md5"
	"errors"
	"fmt"

	"gitlab.com/gitlab-org/gitlab-runner/executors/docker/internal/volumes/parser"
)

var (
	errDirectoryNotAbsolute = errors.New("build directory needs to be an absolute path")
	errDirectoryIsRootPath  = errors.New("build directory needs to be a non-root path")
)

type debugLogger interface {
	Debugln(args ...interface{})
}

func IsHostMountedVolume(volumeParser parser.Parser, dir string, volumes ...string) (bool, error) {
	if !volumeParser.Path().IsAbs(dir) {
		return false, errDirectoryNotAbsolute
	}

	if volumeParser.Path().IsRoot(dir) {
		return false, errDirectoryIsRootPath
	}

	for _, volume := range volumes {
		parsedVolume, err := volumeParser.ParseVolume(volume)
		if err != nil {
			return false, err
		}

		if parsedVolume.Len() < 2 {
			continue
		}

		if volumeParser.Path().Contains(parsedVolume.Destination, dir) {
			return true, nil
		}
	}
	return false, nil
}

func hashContainerPath(containerPath string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(containerPath)))
}

type ErrVolumeAlreadyDefined struct {
	containerPath string
}

func (e *ErrVolumeAlreadyDefined) Error() string {
	return fmt.Sprintf("volume for container path %q is already defined", e.containerPath)
}

func NewErrVolumeAlreadyDefined(containerPath string) *ErrVolumeAlreadyDefined {
	return &ErrVolumeAlreadyDefined{
		containerPath: containerPath,
	}
}

type pathList map[string]bool

func (m pathList) Add(containerPath string) error {
	if m[containerPath] {
		return NewErrVolumeAlreadyDefined(containerPath)
	}

	m[containerPath] = true

	return nil
}
