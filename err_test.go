package alephzero

import (
	"testing"
)

func check(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func TestErr(t *testing.T) {
	err := errorFrom(A0_ERR_AGAIN)
	want := "Not available yet"
	if err.Error() != want {
		t.Errorf("errorFrom(A0_ERR_AGAIN) = %v, want %v", err, want)
	}
}
