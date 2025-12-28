package target

import (
	"github.com/kunalvirwal/shogun-cd/internal/git"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

type Kind string

const TargetKind Kind = "Target"

const (
	ServerType  = "server"
	ClusterType = "cluster"
)

type TargetService interface {
	// Loads a pipeline from the specified yaml file path
	LoadTarget(f []byte) *Target
}

type Service struct {
	logger     utils.Logger
	gitService git.GitService
}

type Target struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       Kind     `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type Metadata struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // server or cluster
}

type Spec struct {
	Host         string `yaml:"host"`
	User         string `yaml:"user"`
	Port         int    `yaml:"port"`
	AccessSecret string `yaml:"access-key-secret"`
}

func NewTargetService(logger utils.Logger, gitService git.GitService) TargetService {
	return &Service{
		logger:     logger,
		gitService: gitService,
	}
}
