package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/appconfigdata"
)

var verbose bool

type AppConfigParams struct {
	applicationID   string
	environmentID   string
	configProfileID string
}

func main() {
	params := readFlags()

	vars, err := getConfig(params)
	if err != nil {
		fmt.Printf("failed to get config: %s\n", err)
		os.Exit(1)
	}

	args := flag.Args()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(), vars...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if verbose {
		fmt.Printf("running %q with args: %s and env: %s\n", args[0], args[1:], cmd.Env)
	}
	if err = cmd.Run(); err != nil {
		os.Exit(2)
	}
}

func readFlags() (params AppConfigParams) {
	flag.StringVar(&params.applicationID, "app", "", "application identifier")
	flag.StringVar(&params.environmentID, "env", "", "environment identifier")
	flag.StringVar(&params.configProfileID, "config", "", "config profile identifier")
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.Parse()

	if params.applicationID == "" {
		fmt.Println("specify application identifier with --app flag")
		os.Exit(1)
	}
	if params.environmentID == "" {
		fmt.Println("specify environment identifier with --env flag")
		os.Exit(1)
	}
	if params.configProfileID == "" {
		fmt.Println("specify config profile identifier with --config flag")
		os.Exit(1)
	}

	if flag.Arg(0) == "" {
		fmt.Println("must specify program to execute")
		os.Exit(1)
	}

	fmt.Printf("reading from AppConfig app %q, env %q, config profile %q\n",
		params.applicationID, params.environmentID, params.configProfileID)

	return
}

func getConfig(params AppConfigParams) ([]string, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := appconfigdata.NewFromConfig(cfg)
	session, err := client.StartConfigurationSession(ctx, &appconfigdata.StartConfigurationSessionInput{
		ApplicationIdentifier:          &params.applicationID,
		ConfigurationProfileIdentifier: &params.configProfileID,
		EnvironmentIdentifier:          &params.environmentID,
	})
	if err != nil {
		return nil, err
	}

	configuration, err := client.GetLatestConfiguration(ctx, &appconfigdata.GetLatestConfigurationInput{
		ConfigurationToken: session.InitialConfigurationToken,
	})
	if err != nil {
		return nil, err
	}

	vars := getVars(string(configuration.Configuration))

	fmt.Printf("read %d lines from AppConfig\n", len(vars))
	if verbose {
		fmt.Printf("vars: %s\n", vars)
	}

	return vars, nil
}

func getVars(config string) []string {
	lines := strings.Split(config, "\n")

	var vars []string
	for _, l := range lines {
		if parsed := parseLine(l); parsed != "" {
			vars = append(vars, l)
		}
	}

	return vars
}

func parseLine(line string) string {
	if len(line) == 0 || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
		return ""
	}

	// strip quotes if present
	split := strings.Split(line, "=")
	key := split[0]
	value := split[1]
	if value[0:1] == `"` && value[len(value)-1:] == `"` {
		line = key + "=" + value[1:len(value)-1]
	}
	return line
}
