package target

import (
	"os"

	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"go.yaml.in/yaml/v3"
)

func (s *Service) LoadTarget(path string) *Target {
	data, err := os.ReadFile(path)
	if err != nil {
		s.logger.LogNewError("failed to read target file: %v", err)
		return nil
	}

	var target Target
	if err := yaml.Unmarshal(data, &target); err != nil {
		s.logger.LogNewError("failed to unmarshal target YAML: %v", err)
		return nil
	}

	// [TODO]: Validate the target structure here if needed
	if !validateTarget(s.logger, &target) {
		return nil
	}

	s.logger.LogInfo("Target loaded: Name=%s, Type=%s, Host=%s", target.Metadata.Name, target.Metadata.Type, target.Spec.Host)
	// [TODO]: Further processing of the loaded target, store into DB and all

	return &target
}

func validateTarget(logger utils.Logger, target *Target) bool {
	if target.ApiVersion != "shogun/v1" {
		logger.LogNewError("invalid apiVersion: %s", target.ApiVersion)
		return false
	}
	if target.Kind != TargetKind {
		logger.LogNewError("invalid kind: %s", target.Kind)
		return false
	}
	if target.Metadata.Type != ServerType && target.Metadata.Type != ClusterType {
		logger.LogNewError("invalid target type: %s", target.Metadata.Type)
		return false
	}
	if target.Spec.Host == "" || target.Spec.User == "" || target.Spec.AccessSecret == "" {
		logger.LogNewError("host, user, and access-key-secret must be provided")
		return false
	}
	if target.Spec.Port <= 0 || target.Spec.Port > 65535 {
		logger.LogNewError("invalid port number: %d", target.Spec.Port)
		return false
	}
	return true
}
