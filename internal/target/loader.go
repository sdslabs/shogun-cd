package target

import (
	"strings"

	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"go.yaml.in/yaml/v3"
)

func (s *Service) LoadTarget(f []byte) *Target {

	var target Target
	if err := yaml.Unmarshal(f, &target); err != nil {
		s.logger.LogNewError("failed to unmarshal target YAML: %v", err)
		return nil
	}

	// Trim whitespace from all string fields
	target.Metadata.Name = strings.TrimSpace(target.Metadata.Name)
	target.Metadata.Type = strings.TrimSpace(target.Metadata.Type)
	target.Spec.Host = strings.TrimSpace(target.Spec.Host)
	target.Spec.User = strings.TrimSpace(target.Spec.User)
	target.Spec.AccessSecret = strings.TrimSpace(target.Spec.AccessSecret)

	// [TODO]: Validate the target structure here if needed
	if !validateTarget(s.logger, &target) {
		return nil
	}

	s.logger.LogInfo("Target loaded: Name=%s, Type=%s, Host=%s", target.Metadata.Name, target.Metadata.Type, target.Spec.Host)
	// [TODO]: Further processing of the loaded target, store into DB and all

	return &target
}

func validateTarget(logger utils.Logger, target *Target) bool {
	if target.ApiVersion != "shogun.dev/v1" {
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
	if target.Metadata.Name == "" {
		logger.LogNewError("target name must be provided")
		return false
	}
	return true
}
