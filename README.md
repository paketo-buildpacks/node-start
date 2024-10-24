# Paketo Buildpack for Node Start
## `gcr.io/paketo-buildpacks/node-start`

The Paketo Node Start CNB sets the start command for a given node application.
The buildpack will detect when a valid javascript file is present within the 
application source code directory (see [Application Detection](#application-detection).

## Integration

This CNB writes a start command, so there's currently no scenario we can
imagine that you would need to require it as dependency. You may include additional
functionality in a separate buildpack without requiring anything from this buildpack.

## Usage

To package this buildpack for consumption:

```shell
$ ./scripts/package.sh --version <version-number>
```

This will create a `build/buildpackage.cnb` file which you can use to build your app as follows:
```shell
pack build <app-name> \
  --path <path-to-app> \
  --buildpack <path/to/node-engine.cnb> \
  --buildpack build/buildpackage.cnb
```

## Graceful shutdown and signal handling

You can add signal handlers in your app to support graceful shutdown and
program interrupts. This buildpack runs the node server as the init process,
and thus it ignores any signal with the default action. As a result, the
process will not terminate on `SIGINT` or `SIGTERM` unless it is coded to do
so. You can also use docker's `--init` flag to wrap your node process with an
init system that will properly handle signals.

## Specifying a project path

To specify a project subdirectory to be used as the root of the app, please use
the `BP_NODE_PROJECT_PATH` environment variable at build time either directly
(e.g. `pack build my-app --env BP_NODE_PROJECT_PATH=./src/my-app`) or through a
[`project.toml` file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md).
This could be useful if your app is a part of a monorepo.

## Application Detection

This buildpack searches your application root for the following files:
1. `server.[c|m]js`
1. `app.[c|m]js`
1. `main.js`
1. `main.cjs`
1. `main.mjs`
1. `index.js`
1. `index.cjs`
1. `index.mjs`
If you have multiple of the above files in your application root then the
highest priority file, for example (`server.js > app.js > main.js > index.js`), will be
chosen for the start command.

## BP_LAUNCHPOINT

The `BP_LAUNCHPOINT` environment variable may be used to specify a file for the
start command that is not included in the above set.

e.g. If `BP_LAUNCHPOINT=./src/launchpoint.js`, the buildpack will verify that
the file exists and then set the start command using that file `node src/launchpoint.js`

## Enabling reloadable process types

You can configure this buildpack to wrap the entrypoint process of your app
such that it kills and restarts the process whenever files in the app's working
directory in the container change. With this feature enabled, copying new
versions of source code into the running container will restart the application process.

Set the environment variable `BP_LIVE_RELOAD_ENABLED=true` at build time to enable this feature.

```shell
pack build my-app \
  --env BP_LIVE_RELOAD_ENABLED=true
````

## Run Tests

To run all unit tests, run:
```shell
./scripts/unit.sh
```

To run all integration tests, run:
```shell
/scripts/integration.sh
```
