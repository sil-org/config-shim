package main

import (
	"reflect"
	"testing"
)

func Test_getVars(t *testing.T) {
	tests := []struct {
		name   string
		config string
		want   []string
	}{
		{
			name:   "empty",
			config: "",
			want:   nil,
		},
		{
			name:   "no value",
			config: "A=",
			want:   []string{"A="},
		},
		{
			name:   "one var",
			config: "A=B",
			want:   []string{"A=B"},
		},
		{
			name:   "two vars",
			config: "A=B\nC=D",
			want:   []string{"A=B", "C=D"},
		},
		{
			name:   "two vars separated by a comment",
			config: "A=B\n# comment\nC=D",
			want:   []string{"A=B", "C=D"},
		},
		{
			name:   "one var with equals in value",
			config: `A="abc123="`,
			want:   []string{`A=abc123=`},
		},
		{
			name:   "everything",
			config: "A=B\n# comment\nC=D=\nE=\"=F\"\n#G=H",
			want:   []string{"A=B", "C=D=", "E==F"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getVars(tt.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "empty",
			line: "",
			want: "",
		},
		{
			name: "no value",
			line: "A=",
			want: "A=",
		},
		{
			name: "comment",
			line: "# comment",
			want: "",
		},
		{
			name: "no =",
			line: "FOO",
			want: "",
		},
		{
			name: "no quotes",
			line: "A=B",
			want: "A=B",
		},
		{
			name: "with quotes",
			line: `A="B C"`,
			want: "A=B C",
		},
		{
			name: "contains equals",
			line: `A=abc123=`,
			want: "A=abc123=",
		},
		{
			name: "with quotes and contains equals",
			line: `A="abc123="`,
			want: "A=abc123=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseLine(tt.line); got != tt.want {
				t.Errorf("parseLine() = %v, want %v", got, tt.want)
			}
		})
	}
}
