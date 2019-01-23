# waitfor

waitfor (or wfor) waits until some condition is fulfilled and then either execs a specified binary or exits with returncode 0. Currently the following things can be waited for:

- `-path PATH`: Wait until `PATH` can be sucessfully `stat`-ed
- `-http URL`: Wait until a request to `URL` completes with a 2xx status code
- `-tcp HOSTPORT`: Wait until a connection to `HOSTPORT` can be established. The connection timeout is currently hard-coded to 500ms
- `-udp HOSTPORT`: Wait until at least one byte could be read from `HOSTPORT`. The read timeout is currently hard-coded to 500ms

Any number of these conditions can be specified. Use the `-and` / `-or` flags to control how they will be evaluated. `-and` / `-or` affect all conditions, so you can either wait until all conditions or any one of them are fulfilled. By default conditions are and-ed.

Anything after `--` on the commandline is interpreted as a binary (including parameters) that will be executed (as in using the `exec` syscall) after waiting finishes successfully.

`-interval` controls how often the conditions are checked, default is 10s. `-timeout` controls the amount of time after which `wfor` will stop checking, default is 5m.
Both parameters use the golang [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) syntax.

The `-debug` parameter enables verbose logging.

## Examples
```bash
# Start nginx after the database is up (path exists AND port is open)
wfor -path /tmp/db.dat -tcp localhost:3306 -- nginx -g 'daemon off;'

# or alternatively
wfor -path /tmp/db.dat -tcp localhost:3306 && nginx -g 'daemon off;'
```

