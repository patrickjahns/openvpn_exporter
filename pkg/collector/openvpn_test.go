package collector

import "testing"

var containsTestCases = []struct {
	scenarioName string
	list         []string
	element      string
	expected     bool
}{
	{"contains element", []string{"bar", "foo"}, "bar", true},
	{"does not contain element", []string{"foo", "bar"}, "baz", false},
}

func TestContainsFunction(t *testing.T) {
	for _, tt := range containsTestCases {
		t.Run(tt.scenarioName, func(t *testing.T) {
			if contains(tt.list, tt.element) != tt.expected {
				t.Errorf("Unexpected result")
			}
		})
	}
}
