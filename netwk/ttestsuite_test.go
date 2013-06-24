package netwk

import (
	"github.com/fmstephe/matching_engine/matcher"
	"testing"
)

var mkr = newMatchTesterMaker()

func TestRunTestSuite(t *testing.T) {
	matcher.RunTestSuite(t, newMatchTesterMaker())
}
