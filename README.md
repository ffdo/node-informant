# Node informant [![Build Status](https://travis-ci.org/dereulenspiegel/node-collector.svg?branch=master)](https://travis-ci.org/dereulenspiegel/node-collector)

Node informant actually consists of two tools. neighbour-discovery and gluon-collector.

This is a little utility to continuously request data from announced enabled
nodes. Currently it uses [gb](http://getgb.io/) as build system. Once the go
vendor feature is not experimental any more (or at least a used default) this
project will move to the default go tools most probably.

## How to build
Simply check out the project and install `gb`. Then simply execute `gb build`
and executables are build. gluon-collector is a daemon for continuous data
collection,neighbour-discovery is replacement for gluon-neighbour-discovery and
executes single queries.

# neighbour-discovery

This tool can act as a replacement for the gluon-neighbour-discovery tool. It has
the following command line switches:

Switch | Description | Default | Mandatory
------ | ----------- | ------- | ---------
-iface | The interface to use to send packets from and receive packets on | none | Yes
-query | The query to send i.e. "GET nodeinfos" | none | Yes
-deflate | Whether to decompress received data via deflate | false | No
-port | The port to bind to, to receive packets on | 12444 | No
-timeout | After how many seconds the program should terminate. -1 to keep it running indefinitely. | -1 | No
-target | If a target IPv6 address is specified, the query is send via unicast to this target | none | No

# gluon-collector

gluon-collector should run in the background. It queries in regular intervals all nodes
listening on the default announced multicast group and stores the received information.
The received information is the available via a REST API or data prepared for meshviewer.

## Command line switches

Switch | Description | Default | Mandatory
------ | ----------- | ------- | ---------
-config | The path to a valid yaml or json config | /etc/node-collector.yaml | No
-import | Import data from this path. The type of data depends on the import type | none | No
-importType | Specify the type of data to import. Currently only ffmap-backend is supported | ffmap-backend | No

Please not that it is not advised to add the import flags to the default startup config,
since this would import the legacy data on every startup, effectively overwriting previously
collected data.

## Example config

```yaml
announced:                # Block describing the behavior of the annoucned requester
  interval:
    statistics: 300       # The interval in seconds to fetch fast changing data like statistics and neighbours
    nodeinfo: 1800        # The interval in seconds to request more static data and discover new nodes
  interface: "bat0"       # The interface to use for announced
  port: 21444             # The port to use as a source port announced requests and to listen for responses on

logger:     
  level: "warn"           # The log level, see logrus for valid values
  file: /var/log/gluon-collector.log  # If the log file is specified the log is written there. If not everything is send to stdout.

store:
  type: "bolt"            # The type of data store to use. Currently bolt (persistend) and memory (non persistend) are supported
  path: "/opt/gluon-collector/collector.db" # The path is only relevant for bolt store. Where to store the database?

http:             
  port: 8079              # The port where the http server will listen on.
  address: "[::]"         # Optional listen address if you want the server to listen only on a specific interface
```
