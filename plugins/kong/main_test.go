package main

import (
	"testing"
)

// In a real scenario, testing a Kong Go PDK plugin thoroughly requires spinning up
// a kong test container or heavily mocking the PDK context.
// For now, we provide a sanity check test.

func TestPluginInitialization(t *testing.T) {
	plugin := New()
	if plugin == nil {
		t.Fatal("expected plugin to be initialized")
	}

	_, ok := plugin.(*OathMeshPlugin)
	if !ok {
		t.Fatal("expected plugin to be of type *OathMeshPlugin")
	}
}

func TestPluginMetadata(t *testing.T) {
	if Version == "" {
		t.Fatal("expected Version to be set")
	}
	if Priority <= 0 {
		t.Fatal("expected Priority to be greater than 0")
	}
}
