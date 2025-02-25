package main

import (
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

func TestGetVars(t *testing.T) {
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
			config: `A=abc123=`,
			want:   []string{"A=abc123="},
		},
		{
			name:   "one var with equals in value, surrounded by quotes",
			config: `A="abc123="`,
			want:   []string{"A=abc123="},
		},
		{
			name:   "comment after value",
			config: `A=abc123 # comment`,
			want:   []string{"A=abc123"},
		},
		{
			name:   "everything",
			config: "A=B\n# comment\nC=D=\nE=\"=F\"\n#G=H",
			want:   []string{"A=B", "C=D=", "E==F"},
		},
		{
			name:   "json with escaped quote marks",
			config: `A="{\"type\":\"service_account\"}"`,
			want:   []string{`A={"type":"service_account"}`},
		},
		{
			name:   "json with unescaped quote marks",
			config: `A={"type":"service_account"}`,
			want:   []string{`A={"type":"service_account"}`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getVars([]byte(tt.config))
			if err != nil {
				t.Errorf("getVars returned err %v, want nil", err)
			}
			slices.Sort(got)
			slices.Sort(tt.want)
			if !slices.Equal(got, tt.want) {
				t.Errorf("getVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceConfigValues(t *testing.T) {
	savedGOOS := os.Getenv("GOOS")
	defer func() {
		err := os.Setenv("GOOS", savedGOOS)
		if err != nil {
			panic(err)
		}
	}()

	err := os.Setenv("GOOS", "linux")
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name    string
		line    string
		want    string
		wantErr bool
	}{
		{
			name:    "no update, no comment",
			line:    "GOOS=windows\nGOARCH=amd64\n",
			want:    "GOOS=windows\nGOARCH=amd64\n",
			wantErr: false,
		},
		{
			name:    "simple update",
			line:    "GOOS=windows # {update}\nGOARCH=amd64\n",
			want:    "GOOS='linux' # {update}\nGOARCH=amd64\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := replaceConfigValues([]byte(tt.line))
			if (err != nil) != tt.wantErr {
				t.Errorf("replaceLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("replaceLine() got = '%v', want '%v'", got, tt.want)
			}
		})
	}
}

func TestReplaceLine(t *testing.T) {
	savedGOOS := os.Getenv("GOOS")
	defer func() {
		err := os.Setenv("GOOS", savedGOOS)
		if err != nil {
			panic(err)
		}
	}()

	err := os.Setenv("GOOS", "linux")
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name     string
		line     string
		variable string
		newValue string
		want     string
		wantErr  bool
	}{
		{
			name:     "no update, no comment",
			line:     "GOOS=windows",
			variable: "GOOS",
			newValue: "linux",
			want:     "GOOS=windows",
			wantErr:  false,
		},
		{
			name:     "no update, preserve the comment",
			line:     "GOOS=windows # GOOS is the target OS",
			variable: "GOOS",
			newValue: "linux",
			want:     "GOOS=windows # GOOS is the target OS",
			wantErr:  false,
		},
		{
			name:     "simple update",
			line:     "GOOS=windows # {update}",
			variable: "GOOS",
			newValue: "linux",
			want:     "GOOS='linux' # {update}",
			wantErr:  false,
		},
		{
			name:     "update with other comment characters",
			line:     "GOOS=windows # GOOS is the target OS {update} it should be replaced with 'linux'",
			variable: "GOOS",
			newValue: "linux",
			want:     "GOOS='linux' # GOOS is the target OS {update} it should be replaced with 'linux'",
			wantErr:  false,
		},
		{
			name:     "update with quoted value",
			line:     "GOOS='windows' # {update}",
			variable: "GOOS",
			newValue: "linux",
			want:     "GOOS='linux' # {update}",
			wantErr:  false,
		},
		{
			name:     "{update} actually in the value should be ignored",
			line:     "GOOS='windows{update}'",
			variable: "GOOS",
			newValue: "linux",
			want:     "GOOS='windows{update}'",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := replaceLine(tt.line, tt.variable, tt.newValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("replaceLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("replaceLine() got = '%v', want '%v'", got, tt.want)
			}
		})
	}
}

func TestGetVarsFromParameters(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		parameters []types.Parameter

		want []string
	}{
		{
			name: "a",
			path: "/path/",
			parameters: []types.Parameter{
				{Name: aws.String("/path/foo"), Value: aws.String("foo_value")},
				{Name: aws.String("/path/bar"), Value: aws.String("bar_value")},
				{Name: nil, Value: aws.String("nil_name")},
				{Name: aws.String("/path/nil_value"), Value: nil},
			},
			want: []string{"foo=foo_value", "bar=bar_value"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getVarsFromParameters(tt.path, tt.parameters); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getVarsFromParameters() = %v, want %v", got, tt.want)
			}
		})
	}
}
