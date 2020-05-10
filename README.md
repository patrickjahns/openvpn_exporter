# openvpn_exporter

[![Release](https://img.shields.io/github/v/release/patrickjahns/openvpn_exporter?sort=semver)](https://github.com/patrickjahns/openvpn_exporter/releases)
[![LICENSE](https://img.shields.io/github/license/patrickjahns/openvpn_exporter)](https://github.com/patrickjahns/openvpn_exporter/blob/master/LICENSE)
[![Test and Build](https://github.com/patrickjahns/openvpn_exporter/workflows/Test%20and%20Build/badge.svg)](https://github.com/patrickjahns/openvpn_exporter/actions?query=workflow%3A%22Test+and+Build%22)
[![Go Doc](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/patrickjahns/openvpn_exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/patrickjahns/openvpn_exporter)](https://goreportcard.com/report/github.com/patrickjahns/openvpn_exporter)

Prometheus exporter for openvpn. Exposes the metrics from [openvpn status file](https://openvpn.net/community-resources/reference-manual-for-openvpn-2-4/) - supports status-version 1-3


## Installation

For pre built binaries, please take a look at the [github releases](https://github.com/patrickjahns/openvpn_exporter/releases)

## Usage

```shell script
$ ./bin/openvpn_exporter -h

GLOBAL OPTIONS:
   --web.address value, --web.listen-address value  Address to bind the metrics server (default: "0.0.0.0:9176") [$OPENVPN_EXPORTER_WEB_ADDRESS]
   --web.path value, --web.telemetry-path value     Path to bind the metrics server (default: "/metrics") [$OPENVPN_EXPORTER_WEB_PATH]
   --web.root value                                 Root path to exporter endpoints (default: "/") [$OPENVPN_EXPORTER_WEB_ROOT]
   --status-file value                              The OpenVPN status file(s) to export (example test:./example/version1.status ) [$OPENVPN_EXPORTER_STATUS_FILE]
   --disable-client-metrics                         Disables per client (bytes_received, bytes_sent, connected_since) metrics (default: false) [$OPENVPN_EXPORTER_DISABLE_CLIENT_METRICS]
   --enable-golang-metrics                          Enables golang and process metrics for the exporter)  (default: false) [$OPENVPN_EXPORTER_ENABLE_GOLANG_METRICS]
   --log.level value                                Only log messages with given severity (default: "info") [$OPENVPN_EXPORTER_LOG_LEVEL]
   --help, -h                                       Show help (default: false)
   --version, -v                                    Prints the current version (default: false)
```

### Example metrics

```
# HELP openvpn_build_info A metric with a constant '1' value labeled by version information
# TYPE openvpn_build_info gauge
openvpn_build_info{date="20200503",go="go1.14.2",revision="f84a7a5",version="f84a7a5"} 1
# HELP openvpn_bytes_received Amount of data received via the connection
# TYPE openvpn_bytes_received gauge
openvpn_bytes_received{common_name="test1@localhost",server="v2"} 3871
openvpn_bytes_received{common_name="test1@localhost",server="v3"} 3871
openvpn_bytes_received{common_name="test@localhost",server="v2"} 3860
openvpn_bytes_received{common_name="test@localhost",server="v3"} 3860
openvpn_bytes_received{common_name="user1",server="v1"} 7.883858e+06
openvpn_bytes_received{common_name="user2",server="v1"} 1.6732e+06
openvpn_bytes_received{common_name="user3@test.de",server="v1"} 1.9602844e+07
openvpn_bytes_received{common_name="user4",server="v1"} 582207
# HELP openvpn_bytes_sent Amount of data sent via the connection
# TYPE openvpn_bytes_sent gauge
openvpn_bytes_sent{common_name="test1@localhost",server="v2"} 3924
openvpn_bytes_sent{common_name="test1@localhost",server="v3"} 3924
openvpn_bytes_sent{common_name="test@localhost",server="v2"} 3688
openvpn_bytes_sent{common_name="test@localhost",server="v3"} 3688
openvpn_bytes_sent{common_name="user1",server="v1"} 7.76234e+06
openvpn_bytes_sent{common_name="user2",server="v1"} 2.065632e+06
openvpn_bytes_sent{common_name="user3@test.de",server="v1"} 2.3599532e+07
openvpn_bytes_sent{common_name="user4",server="v1"} 575193
# HELP openvpn_collection_error Error occured during collection
# TYPE openvpn_collection_error counter
openvpn_collection_error{server="wrong"} 5
# HELP openvpn_connected_since Unixtimestamp when the connection was established
# TYPE openvpn_connected_since gauge
openvpn_connected_since{common_name="test1@localhost",server="v2"} 1.58825494e+09
openvpn_connected_since{common_name="test1@localhost",server="v3"} 1.58825494e+09
openvpn_connected_since{common_name="test@localhost",server="v2"} 1.588254938e+09
openvpn_connected_since{common_name="test@localhost",server="v3"} 1.588254938e+09
openvpn_connected_since{common_name="user1",server="v1"} 1.587551802e+09
openvpn_connected_since{common_name="user2",server="v1"} 1.587551812e+09
openvpn_connected_since{common_name="user3@test.de",server="v1"} 1.587552165e+09
openvpn_connected_since{common_name="user4",server="v1"} 1.587551814e+09
# HELP openvpn_connections Amount of currently connected clients
# TYPE openvpn_connections gauge
openvpn_connections{server="v1"} 4
openvpn_connections{server="v2"} 2
openvpn_connections{server="v3"} 2
# HELP openvpn_last_updated Unix timestamp when the last time the status was updated
# TYPE openvpn_last_updated gauge
openvpn_last_updated{server="v1"} 1.587665671e+09
openvpn_last_updated{server="v2"} 1.588254944e+09
openvpn_last_updated{server="v3"} 1.588254944e+09
# HELP openvpn_max_bcast_mcast_queue_len MaxBcastMcastQueueLen of the server
# TYPE openvpn_max_bcast_mcast_queue_len gauge
openvpn_max_bcast_mcast_queue_len{server="v1"} 5
openvpn_max_bcast_mcast_queue_len{server="v2"} 0
openvpn_max_bcast_mcast_queue_len{server="v3"} 0
# HELP openvpn_server_info A metric with a constant '1' value labeled by version information
# TYPE openvpn_server_info gauge
openvpn_server_info{arch="unknown",server="v1",version="unknown"} 1
openvpn_server_info{arch="x86_64-pc-linux-gnu",server="v2",version="2.4.4"} 1
openvpn_server_info{arch="x86_64-pc-linux-gnu",server="v3",version="2.4.4"} 1
# HELP openvpn_start_time Unix timestamp of the start time of the exporter
# TYPE openvpn_start_time gauge
openvpn_start_time 1.588506393e+09
```

## Development

This project requires Go >= v1.14. To get started, clone the repository and run `make generate`

```shell script
git clone https://github.com/patrickjahns/openvpn_exporter.git
cd openvpn_exporter

make generate
make build

./cmd/openvpn_exporter -h
```

### Building the project

```shell script
make build
```

### Running tests

```shell script
make test
```





