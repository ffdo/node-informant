# Node informant [![Build Status](https://travis-ci.org/ffdo/node-informant.svg?branch=master)](https://travis-ci.org/ffdo/node-informant)

Node informant actually consists of two tools. neighbour-discovery and gluon-collector.

This is a little utility to continuously request data from announced enabled
nodes. Currently it uses [gb](http://getgb.io/) as build system. Once the go
vendor feature is not experimental any more (or at least a used default) this
project will move to the default go tools most probably.

## How to build
Simply check out the project and install [gb](http://getgb.io/). Then simply
execute `gb build` and executables are build. gluon-collector is a daemon for
continuous data collection,neighbour-discovery is replacement for
gluon-neighbour-discovery and executes single queries.

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
The received information is available via a REST API or data prepared for meshviewer.

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
receiver:                 # List of receiver receiving informantion from nodes.     
- type: announced         # Type of the receiver. Currently only announced is supported
  interface: "bat0"       # The interface to use for announced
  port: 21444             # The port to use as a source port announced requests and to listen for responses on

interval:
  statistics: 300         # The interval in seconds to fetch fast changing data like statistics and neighbours
  nodeinfo: 1800          # The interval in seconds to request more static data and discover new nodes
  expire: 3               # This is a multiplicator for the statistics interval. A node is considered offline if
                          # statistics interval multiplied by expire seconds have passed since the last response

logger:     
  level: "warn"           # The log level, see logrus for valid values
  file: /var/log/gluon-collector.log  # If the log file is specified the log is written there. If not everything is send to stdout.

store:
  type: "bolt"            # The type of data store to use. Currently bolt (persistend) and memory (non persistend) are supported
  path: "/opt/gluon-collector/collector.db" # The path is only relevant for bolt store. Where to store the database?
  expireNodesAfterDays: 365 # After this amount of days, a node is considered gone and is deleted from the database

http:             
  port: 8079              # The port where the http server will listen on.
  address: "[::]"         # Optional listen address if you want the server to listen only on a specific interface

prometheus:
  namelabel: true         # Label prometheus node statistics with the host name
  sitecodelabel: true     # Label prometheus node statistics with the received site code
```

## HTTP API

The following rest endpoints are available. All endpoints return JSON (or JSON arrays)
and CORS headers. The only valid method is GET.

Endpoint | Description
-------- | -----------
/nodes.json | Generates a valid nodes.json for meshviewer
/graph.json | Generates valid graph data for meshviewer
/nodeinfos/{nodeid} | Retrieves general node information about the node with nodeid
/nodeinfos | Retrieve all available general node information
/statistics/{nodeid} | Retrieve statistics for node
/statistics | Retrieve all available statistics
/neighbours/{nodeid} | Retrieve mesh neighbour information about node
/neighbours | Retrieve all available neighbour information
/nodestatus/{nodeid} | Retrieve status information like Lastseen, Online status etc. for node
/nodestatus | Retrieve all available status information

## Prometheus

Under the /metrics endpoint prometheus metrics are exposed. Currently we expose the
following metrics

Metric | Description
------ | -----------
total_clients | Total count of clients connected to all nodes
meshnodes_total | Count of all known mesh nodes online and offline
total_traffic_rx | Accumulated on mesh nodes received traffic
total_traffic_tx | Accumulated from mesh nodes transmitted traffic
total_traffic_mgmt_rx | Accumulated on mesh nodes received management traffic
total_traffic_mgmt_tx | Accumulated from mesh nodes transmitted management traffic
meshnodes_online_total | Count of currently available mesh nodes
meshnode_traffic_rx | Received traffic on every node labeled with the nodeid and traffic type
meshnode_traffic_tx | Transmitted traffic from every mesh node labeled with the nodeid and traffic type
meshnode_uptime | Uptime of single mesh nodes labeled with the nodeid
meshnode_clients | Client count on mesh nodes labeled with the nodeid

# gluon-emitter

gluon-emitter is a simple tool to respond to announced requests. It is primarily
designed to run on non-embedded hosts like supernodes. Currently it can parse
a json or yaml file containing static information and respond with these information.

## Command line options

Option | Default | Description | Required
-------|---------|-------------|----------
interface | none | Which interface is used to listen for packets | yes
port | 1101 | Port to listen on | no
alias | none | Path to a file containing static information about this node | yes
group | ff02:0:0:0:0:0:2:1001 | Multicast group to join | no
loglevel | info | Log level as defined by logrus to use | no

## Alias file

The alias file can be a json or yaml file containing the information with which to
respond. The top level keys are the possible queries to answer to. Under each top
level must follow the data structure expected by announced for this query

### Example alias yaml file

```yaml
nodeinfo:
  node_id: 04ceefcafe2a
  network:
    mac: "00:50:56:03:57:2b"
    addresses:
    - "2a03:2260:50:5::8"
    - "fe80::6ce:efff:feca:fe2a"
    mesh:
      bat0:
        interfaces:
          tunnel:
          - "04:ce:ef:ca:fe:2a"
    mesh_interfaces:
    - "04:ce:ef:ca:fe:2a"
statistics:
  node_id: 04ceefcafe2a
  uptime: 12345
  gateway: "04:ce:ef:ca:fe:2a"
  clients:
    wifi: 0
    total: 0
```

### TODO

* Collect statistics dynamically
* Discover as much static information as possible automatically
