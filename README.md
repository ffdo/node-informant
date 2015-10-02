# Node informant

This is a little utility to continuously request data from announced enabled
nodes. Currently it uses [gb](http://getgb.io/) as build system.

## How to build
Simply check out the project and install `gb`. Then simply execute `gb build`
and executables are build. gluon-collector is a daemon for continuous data
collection,neighbour-discovery is replacement for gluon-neighbour-discovery and
executes single queries.
