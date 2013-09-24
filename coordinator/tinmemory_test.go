package coordinator

import (
	"testing"
)

func TestGoodNetwork(t *testing.T) {
	testBadNetwork(t, -1, InMemory)
}
