package orchestrator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

type Meta struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

// RunIndexer runs the indexing process for pipelines and targets and stores them in the orchestrator's atomic maps
// Indexer exits if there is any error while indexing the repo
// Indexer aquires a write lock on the git repo while indexing
func (o *orchestrator) RunIndexer() {
	o.logger.LogInfo("Running Indexer...")

	// Take pipline lock prior to repo lock to avoid deadlocks
	// Similar order should be followed in other places where both locks are taken like mutate steps in pipeline execution
	o.mu.Lock()
	defer o.mu.Unlock()

	// Repo RLock so that the repo can't update while indexing
	o.gitService.RLockRepo()
	defer o.gitService.RUnlockRepo()

	pm := make(PipelineMap)
	tm := make(TargetMap)

	root := o.gitService.GetRepoRoot()

	commonParser := func(path string) error {

		f, err := os.ReadFile(path)
		if err != nil {
			o.logger.LogNewError("Failed to read file %s: %v", path, err)
			return nil // ignore invalid yamls
		}

		var m Meta
		if err := yaml.Unmarshal(f, &m); err != nil {
			o.logger.Log("Failed to unmarshal meta from file %s: %v", path, err)
			return nil // ignore invalid yamls
		}

		if m.APIVersion != "shogun.dev/v1" {
			o.logger.Log("Ignoring file %s", path)
			return nil // ignore non-shogun yamls
		}

		switch m.Kind {
		case "Pipeline":
			p := o.pipelineService.LoadPipeline(f)
			if p != nil {
				// Pipelines are uniquely identified by their lowercase names and no duplicates allowed
				name := strings.ToLower(p.Metadata.Name)
				if _, exists := pm[name]; exists {
					return fmt.Errorf("duplicate pipeline name %q", name)
				}
				pm[name] = p
			} else {
				return fmt.Errorf("Invalid pipeline definition %s", path)
			}

		case "Target":
			t := o.targetService.LoadTarget(f)
			if t != nil {
				// Targets are uniquely identified by their lowercase names and no duplicates allowed
				name := strings.ToLower(t.Metadata.Name)
				if _, exists := tm[name]; exists {
					return fmt.Errorf("duplicate target name %q", name)
				}
				tm[name] = t
			} else {
				return fmt.Errorf("Invalid target definition %s", path)
			}
		}

		return nil

	}

	// If any error is encountered in walking or parsing, walking will stop and no change in state will occur
	err := walkRepo(root, commonParser)
	if err != nil {
		o.logger.LogNewError("Error Indexing repo: %v", err)
		return
	}

	// Now I need to just atomically swap the pointers
	o.Pipelines.Store(&pm)
	o.Targets.Store(&tm)

	o.logger.LogInfo("Indexing completed. Pipelines: %d, Targets: %d", len(pm), len(tm))

}

// WalkRepo walks through the repo and uses the provided callback to process yaml files
func walkRepo(root string, callback func(path string) error) error {

	// This handler calls callback for each yaml file found
	handler := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(d.Name())
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		// Read the file and check if it's a Pipeline or Target
		return callback(path)
	}

	return filepath.WalkDir(root, handler)
}
