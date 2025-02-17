package baur

import (
	"errors"

	"github.com/simplesurance/baur/v1/internal/vcs"
	"github.com/simplesurance/baur/v1/pkg/cfg/resolver"
)

const (
	rootVarName      = "{{ .root }}"
	appVarName       = "{{ .appname }}"
	uuidVarname      = "{{ .uuid }}"
	gitCommitVarname = "{{ .gitcommit }}"
)

// defaultAppCfgResolvers returns the default set of resolvers that is applied on application configs.
func defaultAppCfgResolvers(rootPath, appName string, gitCommitFn func() (string, error)) resolver.Resolver {
	return resolver.List{
		&resolver.StrReplacement{Old: appVarName, New: appName},
		&resolver.StrReplacement{Old: rootVarName, New: rootPath},
		&resolver.UUIDVar{Old: uuidVarname},
		&resolver.CallbackReplacement{
			Old: gitCommitVarname,
			NewFunc: func() (string, error) {
				commit, err := gitCommitFn()
				if errors.Is(err, vcs.ErrVCSRepositoryNotExist) {
					return "", errors.New("baur repository is not part of a git repository")
				}

				return commit, err

			},
		},
		&resolver.EnvVar{},
	}
}
