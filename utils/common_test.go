// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package utils

import (
	"reflect"
	"testing"
)

// TestAppendSlices tests AppendSlices
func TestAppendSlices(t *testing.T) {
	tests := []struct {
		name   string
		slices [][]interface{}
		want   []interface{}
	}{
		{
			name:   "AppendSlices with strings",
			slices: [][]interface{}{{"a", "b", "c"}, {"d", "e", "f"}, {"g", "h", "i"}},
			want:   []interface{}{"a", "b", "c", "d", "e", "f", "g", "h", "i"},
		},
		{
			name:   "AppendSlices with ints",
			slices: [][]interface{}{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}},
			want:   []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name:   "AppendSlices with empty slices",
			slices: [][]interface{}{{}, {}, {}},
			want:   []interface{}{},
		},
		{
			name:   "Append identical slices",
			slices: [][]interface{}{{"a", "b", "c"}, {"a", "b", "c"}},
			want:   []interface{}{"a", "b", "c", "a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AppendSlices(tt.slices...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppendSlices() = %v, want %v", got, tt.want)
			}
		})
	}
}
