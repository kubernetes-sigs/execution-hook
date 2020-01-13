/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import "testing"

func TestContains(t *testing.T) {
	testCases := []struct {
		name           string
		list           []string
		key            string
		expectedResult bool
	}{
		{
			name:           "Contains should find key",
			list:           []string{"fox", "dog", "cat"},
			key:            "dog",
			expectedResult: true,
		},
		{
			name:           "Contains should not find key",
			list:           []string{"fox", "dog", "cat"},
			key:            "owl",
			expectedResult: false,
		},
		{
			name:           "Contains in an empty slice",
			list:           []string{},
			key:            "key",
			expectedResult: false,
		},
		{
			name:           "Contains in a nil slice",
			list:           nil,
			key:            "key",
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		actual := Contains(tc.list, tc.key)
		if actual != tc.expectedResult {
			t.Fatalf("%s failed, Want: %t; Got: %t", tc.name, tc.expectedResult, actual)
		}

	}
}

func TestFilter(t *testing.T) {
	testCases := []struct {
		name     string
		list     []string
		filter   string
		expected []string
	}{
		{
			name:     "Filter should remove the filter string from the slice",
			list:     []string{"dog", "cat", "fox", "pig"},
			filter:   "fox",
			expected: []string{"dog", "cat", "pig"},
		},
		{
			name:     "Filter should not remove anything string from the slice",
			list:     []string{"dog", "cat", "fox", "pig"},
			filter:   "owl",
			expected: []string{"dog", "cat", "fox", "pig"},
		},
		{
			name:     "Filter should not remove anything string from an empty slice",
			list:     []string{},
			filter:   "owl",
			expected: []string{},
		},
		{
			name:     "Filter should not remove anything string from a nil slice",
			list:     nil,
			filter:   "owl",
			expected: []string{},
		},
	}
	for _, tc := range testCases {
		actual := Filter(tc.list, tc.filter)
		if len(actual) != len(tc.expected) {
			t.Fatalf("%s failed, Want: %d; Got: %d ", tc.name, len(tc.expected), len(actual))
		}
		for _, e := range tc.expected {
			if !Contains(actual, e) {
				t.Fatalf("%s failed, Want: %q, Got: %q", tc.name, tc.expected, actual)
			}
		}
	}
}
