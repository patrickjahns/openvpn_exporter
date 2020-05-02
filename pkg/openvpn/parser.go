package openvpn

import (
	"bufio"
	"bytes"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// GlobalStats stores global openvpn statistic information
type GlobalStats struct {
	MaxBcastMcastQueueLen int
}

// Client struct store information from openvpn client statistics
type Client struct {
	CommonName     string
	RealAddress    string
	BytesReceived  float64
	BytesSent      float64
	ConnectedSince time.Time
}

// Status reflects all information in a status log
type Status struct {
	ClientList  []Client
	GlobalStats GlobalStats
	UpdatedAt   time.Time
}

type parseError struct {
	s string
}

func (e *parseError) Error() string {
	return e.s
}

const (
	timefmt = "Mon Jan 2 15:04:05 2006"
)

// ParseFile parses a openvpn status log and returns respective stats
func ParseFile(statusfile string) (*Status, error) {
	conn, err := os.Open(statusfile)
	defer conn.Close()
	if err != nil {
		return nil, err
	}
	status, err := parse(bufio.NewReader(conn))
	if err != nil {
		return nil, err
	}
	return status, nil
}

func parseTime(t string) time.Time {
	loc, _ := time.LoadLocation("Local")
	t2, _ := time.ParseInLocation(timefmt, t, loc)
	return t2
}

func parseIP(ip string) string {
	return net.ParseIP(strings.Split(ip, ":")[0]).String()
}

func parse(reader *bufio.Reader) (*Status, error) {
	buf, _ := reader.Peek(19)
	if bytes.HasPrefix(buf, []byte("OpenVPN CLIENT LIST")) {
		return parseStatusV1(reader)
	}
	if bytes.HasPrefix(buf, []byte("TITLE,OpenVPN")) {
		return parseStatusV2(reader)
	}
	return nil, &parseError{"bad status file"}
}


func parseStatusV1(reader *bufio.Reader) (*Status, error) {
	scanner := bufio.NewScanner(reader)
	var lastUpdatedAt time.Time
	var maxBcastMcastQueueLen int
	var clients []Client
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ",")
		if fields[0] == "Updated" && len(fields) == 2 {
			lastUpdatedAt = parseTime(fields[1])
		} else if fields[0] == "Max bcast/mcast queue length" {
			i, err := strconv.Atoi(fields[1])
			if err == nil {
				maxBcastMcastQueueLen = i
			}
		} else if len(fields) == 5 {
			if fields[0] != "Common Name" {
				bytesRec, _ := strconv.ParseFloat(fields[2], 64)
				bytesSent, _ := strconv.ParseFloat(fields[3], 64)
				client := Client{
					CommonName:     fields[0],
					RealAddress:    parseIP(fields[1]),
					BytesReceived:  bytesRec,
					BytesSent:      bytesSent,
					ConnectedSince: parseTime(fields[4]),
				}
				clients = append(clients, client)
			}
		}
	}
	return &Status{
		GlobalStats: GlobalStats{maxBcastMcastQueueLen},
		UpdatedAt:   lastUpdatedAt,
		ClientList:  clients,
	}, nil
}

func parseStatusV2(reader *bufio.Reader) (*Status, error) {
	scanner := bufio.NewScanner(reader)
	var maxBcastMcastQueueLen int
	var lastUpdatedAt time.Time
	var clients []Client
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ",")
		if fields[0] == "TIME" && len(fields) == 3 {
			updatedAtInt, _ := strconv.ParseInt(fields[2], 10, 64)
			lastUpdatedAt = time.Unix(updatedAtInt,0)
		} else if fields[0] == "CLIENT_LIST" {
			bytesRec, _ := strconv.ParseFloat(fields[5], 64)
			bytesSent, _ := strconv.ParseFloat(fields[6], 64)
			connectedSinceInt, _ := strconv.ParseInt(fields[8], 10, 64)
			client := Client{
				CommonName:     fields[1],
				RealAddress:    parseIP(fields[2]),
				BytesReceived:  bytesRec,
				BytesSent:      bytesSent,
				ConnectedSince: time.Unix(connectedSinceInt,0),
			}
			clients = append(clients, client)
		} else if fields[0] == "GLOBAL_STATS" {
			i, err := strconv.Atoi(fields[2])
			if err == nil {
				maxBcastMcastQueueLen = i
			}
		}
	}
	return &Status{
		GlobalStats: GlobalStats{maxBcastMcastQueueLen},
		UpdatedAt:   lastUpdatedAt,
		ClientList:  clients,
	}, nil
}
