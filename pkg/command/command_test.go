package command

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

// constants

// the regexes below describe the expected /metrics page
// things that are not deterministic/easily predictable like start time or random pseudonym are approximated via patterns

const expectedMetricsRegexV1 = `# HELP openvpn_build_info A metric with a constant '1' value labeled by version information
# TYPE openvpn_build_info gauge
openvpn_build_info\{date="",go="go\d+\.\d+\.\d+",revision="",version="\d+\.\d+\.\d+"\} 1
# HELP openvpn_bytes_received Amount of data received via the connection
# TYPE openvpn_bytes_received gauge
openvpn_bytes_received\{common_name="user1",server="test"\} 7\.883858e\+06
openvpn_bytes_received\{common_name="user2",server="test"\} 1\.6732e\+06
openvpn_bytes_received\{common_name="user3@test\.de",server="test"\} 1\.9602844e\+07
openvpn_bytes_received\{common_name="user4",server="test"\} 582207
# HELP openvpn_bytes_sent Amount of data sent via the connection
# TYPE openvpn_bytes_sent gauge
openvpn_bytes_sent\{common_name="user1",server="test"\} 7\.76234e\+06
openvpn_bytes_sent\{common_name="user2",server="test"\} 2\.065632e\+06
openvpn_bytes_sent\{common_name="user3@test\.de",server="test"\} 2\.3599532e\+07
openvpn_bytes_sent\{common_name="user4",server="test"\} 575193
# HELP openvpn_connected_since Unixtimestamp when the connection was established
# TYPE openvpn_connected_since gauge
openvpn_connected_since\{common_name="user1",server="test"\} 1\.587559002e\+09
openvpn_connected_since\{common_name="user2",server="test"\} 1\.587559012e\+09
openvpn_connected_since\{common_name="user3@test\.de",server="test"\} 1\.587559365e\+09
openvpn_connected_since\{common_name="user4",server="test"\} 1\.587559014e\+09
# HELP openvpn_connections Amount of currently connected clients
# TYPE openvpn_connections gauge
openvpn_connections\{server="test"\} 4
# HELP openvpn_last_updated Unix timestamp when the last time the status was updated
# TYPE openvpn_last_updated gauge
openvpn_last_updated\{server="test"\} 1\.587672871e\+09
# HELP openvpn_max_bcast_mcast_queue_len MaxBcastMcastQueueLen of the server
# TYPE openvpn_max_bcast_mcast_queue_len gauge
openvpn_max_bcast_mcast_queue_len\{server="test"\} 5
# HELP openvpn_server_info A metric with a constant '1' value labeled by version information
# TYPE openvpn_server_info gauge
openvpn_server_info\{arch="unknown",server="test",version="unknown"\} 1
# HELP openvpn_start_time Unix timestamp of the start time of the exporter
# TYPE openvpn_start_time gauge
openvpn_start_time \d+\.\d+e\+\d+
`

const expectedMetricsRegexV2 = `# HELP openvpn_build_info A metric with a constant '1' value labeled by version information
# TYPE openvpn_build_info gauge
openvpn_build_info\{date="",go="go\d+\.\d+\.\d+",revision="",version="\d+\.\d+\.\d+"\} 1
# HELP openvpn_bytes_received Amount of data received via the connection
# TYPE openvpn_bytes_received gauge
openvpn_bytes_received\{common_name="test1@localhost",server="test"\} 3871
openvpn_bytes_received\{common_name="test@localhost",server="test"\} 3860
# HELP openvpn_bytes_sent Amount of data sent via the connection
# TYPE openvpn_bytes_sent gauge
openvpn_bytes_sent\{common_name="test1@localhost",server="test"\} 3924
openvpn_bytes_sent\{common_name="test@localhost",server="test"\} 3688
# HELP openvpn_connected_since Unixtimestamp when the connection was established
# TYPE openvpn_connected_since gauge
openvpn_connected_since\{common_name="test1@localhost",server="test"\} 1\.58825494e\+09
openvpn_connected_since\{common_name="test@localhost",server="test"\} 1\.588254938e\+09
# HELP openvpn_connections Amount of currently connected clients
# TYPE openvpn_connections gauge
openvpn_connections\{server="test"\} 2
# HELP openvpn_last_updated Unix timestamp when the last time the status was updated
# TYPE openvpn_last_updated gauge
openvpn_last_updated\{server="test"\} 1\.588254944e\+09
# HELP openvpn_max_bcast_mcast_queue_len MaxBcastMcastQueueLen of the server
# TYPE openvpn_max_bcast_mcast_queue_len gauge
openvpn_max_bcast_mcast_queue_len\{server="test"\} 0
# HELP openvpn_server_info A metric with a constant '1' value labeled by version information
# TYPE openvpn_server_info gauge
openvpn_server_info\{arch="x86_64-pc-linux-gnu",server="test",version="2\.4\.4"\} 1
# HELP openvpn_start_time Unix timestamp of the start time of the exporter
# TYPE openvpn_start_time gauge
openvpn_start_time \d+\.\d+e\+\d+
`

const expectedMetricsRegexV3 = expectedMetricsRegexV2 // the expected output is currently equal in v2 and v3

const expectedPseudonymizedMetricsRegexV1 = `# HELP openvpn_build_info A metric with a constant '1' value labeled by version information
# TYPE openvpn_build_info gauge
openvpn_build_info\{date="",go="go\d+\.\d+\.\d+",revision="",version="\d+\.\d+\.\d+"\} 1
# HELP openvpn_bytes_received Amount of data received via the connection
# TYPE openvpn_bytes_received gauge
openvpn_bytes_received\{common_name="([A-Za-z]+)",server="test"\} \d+(\.\d+e\+\d+){0,1}
openvpn_bytes_received\{common_name="([A-Za-z]+)",server="test"\} \d+(\.\d+e\+\d+){0,1}
openvpn_bytes_received\{common_name="([A-Za-z]+)",server="test"\} \d+(\.\d+e\+\d+){0,1}
openvpn_bytes_received\{common_name="([A-Za-z]+)",server="test"\} \d+(\.\d+e\+\d+){0,1}
# HELP openvpn_bytes_sent Amount of data sent via the connection
# TYPE openvpn_bytes_sent gauge
openvpn_bytes_sent\{common_name="([A-Za-z]+)",server="test"\} \d+(\.\d+e\+\d+){0,1}
openvpn_bytes_sent\{common_name="([A-Za-z]+)",server="test"\} \d+(\.\d+e\+\d+){0,1}
openvpn_bytes_sent\{common_name="([A-Za-z]+)",server="test"\} \d+(\.\d+e\+\d+){0,1}
openvpn_bytes_sent\{common_name="([A-Za-z]+)",server="test"\} \d+(\.\d+e\+\d+){0,1}
# HELP openvpn_connected_since Unixtimestamp when the connection was established
# TYPE openvpn_connected_since gauge
openvpn_connected_since\{common_name="([A-Za-z]+)",server="test"\} \d+(\.\d+e\+\d+){0,1}
openvpn_connected_since\{common_name="([A-Za-z]+)",server="test"\} \d+(\.\d+e\+\d+){0,1}
openvpn_connected_since\{common_name="([A-Za-z]+)",server="test"\} \d+(\.\d+e\+\d+){0,1}
openvpn_connected_since\{common_name="([A-Za-z]+)",server="test"\} \d+(\.\d+e\+\d+){0,1}
# HELP openvpn_connections Amount of currently connected clients
# TYPE openvpn_connections gauge
openvpn_connections\{server="test"\} 4
# HELP openvpn_last_updated Unix timestamp when the last time the status was updated
# TYPE openvpn_last_updated gauge
openvpn_last_updated\{server="test"\} 1\.587672871e\+09
# HELP openvpn_max_bcast_mcast_queue_len MaxBcastMcastQueueLen of the server
# TYPE openvpn_max_bcast_mcast_queue_len gauge
openvpn_max_bcast_mcast_queue_len\{server="test"\} 5
# HELP openvpn_server_info A metric with a constant '1' value labeled by version information
# TYPE openvpn_server_info gauge
openvpn_server_info\{arch="unknown",server="test",version="unknown"\} 1
# HELP openvpn_start_time Unix timestamp of the start time of the exporter
# TYPE openvpn_start_time gauge
openvpn_start_time \d+\.\d+e\+\d+
`

const expectedPseudonymizedMetricsRegexV2 = `# HELP openvpn_build_info A metric with a constant '1' value labeled by version information
# TYPE openvpn_build_info gauge
openvpn_build_info\{date="",go="go\d+\.\d+\.\d+",revision="",version="\d+\.\d+\.\d+"\} 1
# HELP openvpn_bytes_received Amount of data received via the connection
# TYPE openvpn_bytes_received gauge
openvpn_bytes_received\{common_name="([A-Za-z]+)",server="test"\} \d+
openvpn_bytes_received\{common_name="([A-Za-z]+)",server="test"\} \d+
# HELP openvpn_bytes_sent Amount of data sent via the connection
# TYPE openvpn_bytes_sent gauge
openvpn_bytes_sent\{common_name="([A-Za-z]+)",server="test"\} (\d+)
openvpn_bytes_sent\{common_name="([A-Za-z]+)",server="test"\} (\d+)
# HELP openvpn_connected_since Unixtimestamp when the connection was established
# TYPE openvpn_connected_since gauge
openvpn_connected_since\{common_name="([A-Za-z]+)",server="test"\} 1\.\d+e\+09
openvpn_connected_since\{common_name="([A-Za-z]+)",server="test"\} 1\.\d+e\+09
# HELP openvpn_connections Amount of currently connected clients
# TYPE openvpn_connections gauge
openvpn_connections\{server="test"\} 2
# HELP openvpn_last_updated Unix timestamp when the last time the status was updated
# TYPE openvpn_last_updated gauge
openvpn_last_updated\{server="test"\} 1\.588254944e\+09
# HELP openvpn_max_bcast_mcast_queue_len MaxBcastMcastQueueLen of the server
# TYPE openvpn_max_bcast_mcast_queue_len gauge
openvpn_max_bcast_mcast_queue_len\{server="test"\} 0
# HELP openvpn_server_info A metric with a constant '1' value labeled by version information
# TYPE openvpn_server_info gauge
openvpn_server_info\{arch="x86_64-pc-linux-gnu",server="test",version="2\.4\.4"\} 1
# HELP openvpn_start_time Unix timestamp of the start time of the exporter
# TYPE openvpn_start_time gauge
openvpn_start_time \d+\.\d+e\+\d+
`

const expectedPseudonymizedMetricsRegexV3 = expectedPseudonymizedMetricsRegexV2 // the expected output is currently equal in v2 and v3

// helper functions
func version1StatusFile() string {
	statusFileDir := filepath.Join(getCurrentPath(), "..", "..", "example")
	statusFile := filepath.Join(statusFileDir, "version1.status")
	return statusFile
}

func version2StatusFile() string {
	statusFileDir := filepath.Join(getCurrentPath(), "..", "..", "example")
	statusFile := filepath.Join(statusFileDir, "version2.status")
	return statusFile
}

func version3StatusFile() string {
	statusFileDir := filepath.Join(getCurrentPath(), "..", "..", "example")
	statusFile := filepath.Join(statusFileDir, "version3.status")
	return statusFile
}

func getFreeTCPPort() int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port
}

func waitUntilReachable(addr string, retry int) error {
	for i := 0; i < retry; i++ {
		conn, _ := net.DialTimeout("tcp", addr, time.Second)
		if conn != nil {
			conn.Close()
			return nil
		}
	}
	return fmt.Errorf("connection to %s could not be established", addr)
}

func getCurrentPath() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}

func runServer(server *http.Server) {
	_ = server.ListenAndServe()
}

func startCli(t *testing.T) *http.Server {
	app, cfg := initApp()
	var server *http.Server
	app.Action = func(c *cli.Context) error {
		server, _ = run(cfg)
		go runServer(server)
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		t.Fatal(err)
	}
	return server
}

func startWithArgs(args []string, t *testing.T) (string, error) {
	// normalize time zone
	err := os.Setenv("TZ", "UTC")
	if err != nil {
		t.Fatal(err)
	}

	// restore os args after each test to avoid potential issues with the test runner
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	listenAddress := fmt.Sprintf("127.0.0.1:%d", getFreeTCPPort())

	os.Args = append([]string{"openvpn_exporter", "--web.address", listenAddress}, args...)

	server := startCli(t)
	t.Cleanup(func() { server.Close() })

	err = waitUntilReachable(listenAddress, 5)
	if err != nil {
		t.Fatal(err)
	}
	return listenAddress, err
}

func fetchMetrics(t *testing.T, listenAddress string) string {
	client := http.Client{
		Timeout: time.Second,
	}
	resp, err := client.Get(fmt.Sprintf("http://%s/metrics", listenAddress))
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func checkMetrics(t *testing.T, statusFile string, extraArgs []string) string {
	args := append([]string{"--status-file", fmt.Sprintf("test:%s", statusFile)}, extraArgs...)
	listenAddress, err := startWithArgs(args, t)
	if err != nil {
		t.Fatal(err)
	}
	metrics := fetchMetrics(t, listenAddress)
	return metrics
}

func checkMetricsRegex(t *testing.T, statusFile string, expectedMetricsRegex string, extraArgs []string) {
	metrics := checkMetrics(t, statusFile, extraArgs)

	metricsLines := strings.Split(metrics, "\n")
	expectedMetricsRegexLines := strings.Split(expectedMetricsRegex, "\n")

	if len(metricsLines) != len(expectedMetricsRegexLines) {
		t.Fatalf("expectedMetricsRegex must have as many lines as the resulting metrics page: %d - %d", len(metricsLines), len(expectedMetricsRegexLines))
	}

	// check regex line for line to get readable error output
	for idx, regex := range expectedMetricsRegexLines {
		regex = fmt.Sprintf("^%s$", regex)
		assert.Regexp(t, regex, metricsLines[idx])
	}

}

// test functions

func TestRunWithV1Status(t *testing.T) {
	checkMetricsRegex(t, version1StatusFile(), expectedMetricsRegexV1, []string{})
}

func TestRunWithV2Status(t *testing.T) {
	checkMetricsRegex(t, version2StatusFile(), expectedMetricsRegexV2, []string{})
}

func TestRunWithV3Status(t *testing.T) {
	checkMetricsRegex(t, version3StatusFile(), expectedMetricsRegexV3, []string{})
}

func TestRunWithV1StatusAndPseudonymization(t *testing.T) {
	checkMetricsRegex(t, version1StatusFile(), expectedPseudonymizedMetricsRegexV1, []string{"--pseudonymize-client-metrics"})
}

func TestRunWithV2StatusAndPseudonymization(t *testing.T) {
	checkMetricsRegex(t, version2StatusFile(), expectedPseudonymizedMetricsRegexV2, []string{"--pseudonymize-client-metrics"})
}

func TestRunWithV3StatusAndPseudonymization(t *testing.T) {
	checkMetricsRegex(t, version3StatusFile(), expectedPseudonymizedMetricsRegexV3, []string{"--pseudonymize-client-metrics"})
}
