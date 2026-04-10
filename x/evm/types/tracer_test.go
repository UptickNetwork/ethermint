package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewNoOpTracer(t *testing.T) {
	hooks := NewNoOpTracer()
	require.NotNil(t, hooks)
	require.NotNil(t, hooks.OnOpcode)
	require.NotNil(t, hooks.OnFault)
	require.NotNil(t, hooks.OnExit)
	require.NotNil(t, hooks.OnEnter)
}
