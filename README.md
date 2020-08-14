# Node Start Cloud Native Buildpack
## `gcr.io/paketo-buildpacks/node-start`

The Paketo Node Start CNB sets the start command for a given node application.

## Integration

This CNB writes a start command, so there's currently no scenario we can
imagine that you would need to require it as dependency. If a user likes to
include some other functionality, it can be done independent of the Node Start
CNB without requiring a dependency of it.

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
## `buildpack.yml` Configurations

There are no extra configurations for this buildpack based on `buildpack.yml`.
