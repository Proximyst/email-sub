package ids_test

import (
	"testing"

	"github.com/proximyst/email-sub/pkg/ids"
	"github.com/stretchr/testify/require"
)

func TestCalculateID(t *testing.T) {
	t.Parallel()

	id := ids.CalculateID("this is a very long string that is going to be used as a test case for the CalculateID function")
	require.Len(t, id, 41)
}
