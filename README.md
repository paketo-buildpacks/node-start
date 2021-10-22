# Node Start Cloud Native Buildpack
## `gcr.io/paketo-buildpacks/node-start`

The Paketo Node Start CNB sets the start command for a given node application.
The buildpack expects that the app contains a valid `server.js` at the root.

## Integration

This CNB writes a start command, so there's currently no scenario we can
imagine that you would need to require it as dependency. If a user likes to
include some other functionality, it can be done independent of the Node Start
CNB without requiring a dependency of it.

## Usage

To package this buildpack for consumption:

```
$ ./scripts/package.sh --version <version-number>
```

This will create a `buildpackage.cnb` file under the `build` directory which you
can use to build your app as follows:
```
pack build <app-name> -p <path-to-app> -b <path/to/node-engine.cnb> -b build/buildpackage.cnb
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
[`project.toml`
file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md).
This could be useful if your app is a part of a monorepo.

## Application Detection
This buildpack searches your application root for the following files:
1. `server.js`
1. `app.js`
1. `main.js`
1. `index.js`
If you have multiple of the above files in your application root then the
highest priority file (`server.js > app.js > main.js > index.js`) will be
chosen for the start command.

## BP_LAUNCHPOINT

The BP_LAUNCHPOINT environment variable may be used to specify a file for the
start command that is not included in the above set.

e.g. If `BP_LAUNCHPOINT=./src/launchpoint.js`, the buildpack will verify that
the file exists and then set the start command using that file `node
src/launchpoint.js`

## Enabling reloadable process types

You can configure this buildpack to wrap the entrypoint process of your app
such that it kills and restarts the process whenever files in the app's working
directory in the container change. With this feature enabled, copying new
verisons of source code into the running container will trigger your app's
process to restart. Set the environment variable `BP_LIVE_RELOAD_ENABLED=true`
at build time to enable this feature.

## Run Tests

To run all unit tests, run:
```
./scripts/unit.sh
```

To run all integration tests, run:
```
/scripts/integration.sh
```
