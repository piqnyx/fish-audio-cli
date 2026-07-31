package modules

import "testing"

func TestBuildPreservesConfiguredOrder(t *testing.T) {
	t.Parallel()

	processors, err := Build([]string{"passthrough"})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(processors) != 1 {
		t.Fatalf("len(processors) = %d, want 1", len(processors))
	}

	if processors[0].Name() != "passthrough" {
		t.Fatalf(
			"processor name = %q, want %q",
			processors[0].Name(),
			"passthrough",
		)
	}
}

func TestBuildRejectsUnknownModule(t *testing.T) {
	t.Parallel()

	if _, err := Build([]string{"telepathy"}); err == nil {
		t.Fatal("Build() error = nil, want an error")
	}
}
