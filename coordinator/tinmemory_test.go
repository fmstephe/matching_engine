package coordinator

import (
	"testing"
)

func TestGoodNetwork(t *testing.T) {
	testBadNetwork(t, 0.0, InMemory)
}
