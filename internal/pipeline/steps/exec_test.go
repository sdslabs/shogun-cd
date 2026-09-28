package pipelineSteps

import (
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestExecStepShellPathFromYAML(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want string
	}{
		{
			name: "default shell is bash",
			yaml: "exec:\n  target: prod\n  commands:\n    - echo ready\n",
			want: "/bin/bash",
		},
		{
			name: "bash can be selected explicitly",
			yaml: "exec:\n  target: prod\n  shell: bash\n  commands:\n    - echo ready\n",
			want: "/bin/bash",
		},
		{
			name: "sh can be selected",
			yaml: "exec:\n  target: prod\n  shell: sh\n  commands:\n    - echo ready\n",
			want: "/bin/sh",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var wrapper StepWrapper
			if err := yaml.Unmarshal([]byte(test.yaml), &wrapper); err != nil {
				t.Fatalf("unmarshal step: %v", err)
			}
			step, ok := wrapper.Step.(*ExecStep)
			if !ok {
				t.Fatalf("expected ExecStep, got %T", wrapper.Step)
			}

			got, err := step.shellPath()
			if err != nil {
				t.Fatalf("shellPath: %v", err)
			}
			if got != test.want {
				t.Fatalf("shellPath() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestExecStepShellPathRejectsUnsupportedShell(t *testing.T) {
	step := &ExecStep{Shell: "bash -c 'command'"}

	if _, err := step.shellPath(); err == nil {
		t.Fatal("shellPath() accepted unsupported shell")
	}
}
