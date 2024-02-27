package main

import (
	"context"
	"flag"
	"log"
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
		log.Fatalf("failed to get config: %s", err)
	}

	args := flag.Args()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(), vars...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if verbose {
		log.Printf("running %q with args: %s and env: %s", args[0], args[1:], cmd.Env)
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
		log.Printf("specify application identifier with --app flag")
		os.Exit(1)
	}
	if params.environmentID == "" {
		log.Printf("specify environment identifier with --env flag")
		os.Exit(1)
	}
	if params.configProfileID == "" {
		log.Printf("specify config profile identifier with --config flag")
		os.Exit(1)
	}

	if flag.Arg(0) == "" {
		log.Printf("must specify program to execute")
		os.Exit(1)
	}

	if verbose {
		log.Printf("reading from app %q, env %q, profile %q",
			params.applicationID, params.environmentID, params.configProfileID)
	}

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

	if verbose {
		log.Printf("read %d lines", len(vars))
		log.Printf("vars: %s", vars)
	}

	return vars, nil
}

func getVars(config string) []string {
	lines := strings.Split(config, "\n")

	var vars []string
	for _, l := range lines {
		if len(l) == 0 || strings.HasPrefix(l, "#") || !strings.Contains(l, "=") {
			continue
		}

		// strip quotes if present
		if l[0:1] == `"` && l[len(l)-1:] == `"` {
			l = l[1 : len(l)-1]
		}

		vars = append(vars, l)
	}

	return vars
}
