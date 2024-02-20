# config-shim

Languages like PHP, when used in a web application environment are not well-suited for application configuration tools like AWS AppConfig. The entire application is initialized for every web request. This means that without persistent data storage, configuration must be fetched from the API on every request. This tool, provides for a quick migration from an environment-based configuration to AppConfig by reading from the config API once at the time the server starts. It passes configuration into environment variables in the process that starts your web server.

## Configuration

In your startup script insert a call to `config-shim` like this:

```shell
config-shim --app my_app --config default --env prod apache2ctl -D FOREGROUND
```

## Parameters

config-shim command-line parameter format is like `config-shim <flags> <command>`

### Flags
- `--app`: Application Identifier, can be the name of the application or the ID assigned by AWS
- `--config`: Configuration Profile Identifier, can be the profile name or the ID assigned by AWS
- `--env`: Environment Identifier, can be the name of the environment or the ID assigned by AWS

### Command
All parameters after the last flag are used as the command to execute after loading the environment variables from the config data received from AppConfig.
