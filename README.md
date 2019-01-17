# Notes
- a tool that waits until some pre-requisite becomes available and then either exits or executes a specified program
- pre-requisites might be
    - A filesystem path exists
    - TCP/UDP connection succeeds
    - HTTP connection succeeds / specific status
    - HTTP connection yields a specific response

Ideall it's possible to specify all pre-requisites directly on the command line so that no further config files need to be written.

The idea is to delay service startup until all requirements are met to avoid unnecessary container/service restarts

To AND multiple conditions one can chain wfor calls, the last one invoking the desired service.
When specifying multiple conditions, they are OR'd together.

A timeout value can be given after which wfor will fail if no condition has been fulfilled.

# Parameters

# Syntax
--interval to specify the duration to wait between checking if the requisites have been met

wfor --path /some/db/file -- /usr/bin/service --config config.yml --db /some/db/file

wfor --tcp host:port 
wfor --udp host:port

wfor --http host:port --status 200,204
wfor --http host:port --regex '.*"status": "ready".*'


wfor --