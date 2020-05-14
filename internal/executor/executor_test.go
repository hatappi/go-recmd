package executor

import (
	"context"
	"errors"
	"testing"
)

func TestIsIgnoreError(t *testing.T) {
	testCases := []struct {
		err    error
		expect bool
	}{
		{
			err:    nil,
			expect: true,
		},
		{
			err:    errors.New("test"),
			expect: false,
		},
		{
			err:    context.Canceled,
			expect: true,
		},
		{
			err:    context.DeadlineExceeded,
			expect: true,
		},
	}

	for _, tc := range testCases {
		actual := isIgnoreError(tc.err)
		if tc.expect != actual {
			t.Fatalf("failed to match. expect: %v, actual: %v, err: %v", tc.expect, actual, tc.err)
		}
	}
}
