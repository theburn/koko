package auth

import (
	"testing"
)

func TestParseDirectUserFormat(t *testing.T) {

	res := []string{"opt", "SMS"}
	in, qe := CreateChallengerInstruction(res)
	t.Log(in, qe)
}
