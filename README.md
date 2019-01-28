# waitfor

waitfor (or wfor) waits until some condition is fulfilled and then either execs a specified binary or exits with returncode 0. Currently the following things can be waited for:

- `-path PATH`: Wait until `PATH` can be sucessfully `stat`-ed
- `-http SPEC`: Wait until a response according to `SPEC` is received. `SPEC` has the following format ```URL|AcceptedHTTPCodes|Regex```, where `URL` is the URL that the request will be sent to. `AcceptedHTTPCodes` is a list / range of HTTP Status codes that are considered OK by the check. If ommitted any HTTP status will be accepted. `Regex` is a regular expression that will be matched against the response body. If ommitted the response content will not be looked at to determine success. Look at [Examples](https://github.com/pbaettig/waitfor#examples) for further details.
- `-tcp HOSTPORT`: Wait until a connection to `HOSTPORT` can be established. The connection timeout is currently hard-coded to 500ms
- `-udp HOSTPORT`: Wait until at least one byte could be read from `HOSTPORT`. The read timeout is currently hard-coded to 500ms

Any number of these conditions can be specified. Use the `-and` / `-or` flags to control how they will be evaluated. `-and` / `-or` affect all conditions, so you can either wait until all conditions or any one of them are fulfilled. By default conditions are and-ed.

Anything after `--` on the commandline is interpreted as a binary (including parameters) that will be executed (as in using the `exec` syscall) after waiting finishes successfully.

`-interval` controls how often the conditions are checked, default is 10s, minimum is 1s. `-timeout` controls the amount of time after which `wfor` will stop checking, default is 5m.
Both parameters use the golang [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) syntax.

The `-debug` parameter enables verbose logging.


## Examples
```bash
# Start nginx after the database is up (path exists AND port is open)
wfor -path /tmp/db.dat -tcp localhost:3306 -- nginx -g 'daemon off;'

# or alternatively if run in a shell
wfor -path /tmp/db.dat -tcp localhost:3306 && nginx -g 'daemon off;'

# Check URL, accept status 200-208 and 226, require "Authenticated as .*" in the response body
wfor -http "http://some-service.acme.org:8080/secret|200-208,226|Authenticated as .*" && echo "User is authenticated"

# Check URL, accept any status, require "Hello" in response body
wfor -http "http://some-service.acme.org:8080/secret||Hello" && echo "Greeting received"
```
