package netwk

import (
	"github.com/fmstephe/matching_engine/matcher"
	"testing"
)

func TestRunTestSuite(t *testing.T) {
	matcher.RunTestSuite(t, testMkrUtil)
}
