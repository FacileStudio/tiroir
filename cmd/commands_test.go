package cmd

import (
	"reflect"
	"testing"
)

// `tiroir set KEY -value` is the whole point of the store for secrets that look
// like flags. The command never parsed the value as a flag, but pflag stopped at
// the first dash and errored before RunE ran; this pins the interspersed=false
// configuration in init that keeps the value intact.
func TestSetCommandKeepsDashValueAsArgument(t *testing.T) {
	if err := setCmd.ParseFlags([]string{"KEY", "-leading-dash"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	got := setCmd.Flags().Args()
	want := []string{"KEY", "-leading-dash"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %v; want %v", got, want)
	}
}
