package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_chainedCallback(t *testing.T) {
	count := 0
	var add = func() { count++ }
	Chainable().And(add)()
	require.Equal(t, 1, count)

	count = 0
	Chainable(add)()
	require.Equal(t, 1, count)

	count = 0
	Chainable(add).And(add)()
	require.Equal(t, 2, count)

	count = 0
	Chainable(add, add).And(add)()
	require.Equal(t, 3, count)

	Chainable()()
	require.Equal(t, 3, count)
}
