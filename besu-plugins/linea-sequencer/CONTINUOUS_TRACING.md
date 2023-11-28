# Continuous Tracing

The continuous tracing plugin allows to trace every newly imported block and use Corset to check if the constraints are
valid. In case of an error a message will be sent to the configured Slack channel.

## Usage

The continuous tracing plugin is disabled by default. To enable it, use the `--plugin-linea-continuous-tracing-enabled`
flag. If the plugin is enabled it is mandatory to specify the location of `zkevm.bin` using
the `--plugin-linea-continuous-tracing-zkevm-bin` flag. The user with which the node is running needs to have the
appropriate permissions to access `zkevm.bin`.

In order to send a message to Slack a webhook URL needs to be specified by setting the `SLACK_SHADOW_NODE_WEBHOOK_URL`
environment variable. An environment variable was chosen instead of a command line flag to avoid leaking the webhook URL
in the process list.

The environment variable can be set via systemd using the following command:

```shell
Environment=SLACK_SHADOW_NODE_WEBHOOK_URL=https://hooks.slack.com/services/SECRET_VALUES
```

## Invalid trace handling

In the success case the trace file will simply be deleted.

In case of an error the trace file will be renamed to `trace_$BLOCK_NUMBER_$BLOCK_HASH.lt` and moved
to `$BESU_USER_HOME/invalid-traces`. The output of Corset will be saved in the same directory in a file
named `corset_output_$BLOCK_NUMBER_$BLOCK_HASH.txt`. After that an error message will be sent to Slack.
