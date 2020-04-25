package openvpn

import (
	"bufio"
	"strings"
	"testing"
	"time"
)

func TestErrorOnNonExistentFile(t *testing.T) {
	_, e := ParseFile("fixtures/nonExistentFile")
	if e == nil {
		t.Errorf("Parsing Non Existent File failed")
	}
}

func TestParsingEmptyFile(t *testing.T) {
	_, e := parse(bufio.NewReader(strings.NewReader("")))
	if e == nil || e.Error() != "bad status file" {
		t.Errorf("Should have errorred on empty status file")
	}
}

const noConnectedClients = `OpenVPN CLIENT LIST
Updated,Thu Apr 23 20:14:31 2020
Common Name,Real Address,Bytes Received,Bytes Sent,Connected Since
ROUTING TABLE
Virtual Address,Common Name,Real Address,Last Ref
GLOBAL STATS
Max bcast/mcast queue length,0
END
`

func TestNoConnectedClientsAreParsedCorrectly(t *testing.T) {
	status, e := parse(bufio.NewReader(strings.NewReader(noConnectedClients)))
	if e != nil {
		t.Errorf("should have worked")
	}
	loc, _ := time.LoadLocation("Local")
	expectedTime, _ := time.ParseInLocation(timefmt, "Thu Apr 23 20:14:31 2020", loc)
	if !expectedTime.Equal(status.UpdatedAt) {
		t.Errorf("time was not parsed correctly")
	}
	if status.GlobalStats.MaxBcastMcastQueueLen != 0 {
		t.Errorf("MaxBcastMcastQueueLen was not parsed correctly")
	}
}

const connectedClients = `OpenVPN CLIENT LIST
Updated,Thu Apr 23 20:14:31 2020
Common Name,Real Address,Bytes Received,Bytes Sent,Connected Since
user1,1.2.3.4:60102,7883858,7762340,Wed Apr 22 12:36:42 2020
user2,1.2.3.5:50976,1673200,2065632,Wed Apr 22 12:36:52 2020
user3@test.de,1.2.3.6:57688,19602844,23599532,Wed Apr 22 12:42:45 2020
user4,1.2.3.7:40832,582207,575193,Wed Apr 22 12:36:54 2020
ROUTING TABLE
Virtual Address,Common Name,Real Address,Last Ref
10.240.1.222,user4,1.2.3.7:40832,Wed Apr 22 12:36:56 2020
10.240.10.126,user1,1.2.3.4:50976,Thu Apr 23 18:44:56 2020
10.240.79.134,user2,1.2.3.5:60102,Thu Apr 23 20:14:30 2020
10.240.37.214,user3@test.de,1.2.3.6:57688,Thu Apr 23 20:14:16 2020
GLOBAL STATS
Max bcast/mcast queue length,5
END
`

func TestConnectedClientsParsedCorrectly(t *testing.T) {
	status, e := parse(bufio.NewReader(strings.NewReader(connectedClients)))
	if e != nil {
		t.Errorf("should have worked")
	}
	loc, _ := time.LoadLocation("Local")
	expectedTime, _ := time.ParseInLocation(timefmt, "Thu Apr 23 20:14:31 2020", loc)
	if !expectedTime.Equal(status.UpdatedAt) {
		t.Errorf("time was not parsed correctly")
	}
	if status.GlobalStats.MaxBcastMcastQueueLen != 5 {
		t.Errorf("MaxBcastMcastQueueLen was not parsed correctly")
	}
	if len(status.ClientList) != 4 {
		t.Errorf("Clients are not parsed correctly")
	}
	if status.ClientList[0].CommonName != "user1" {
		t.Errorf("Clients are not parsed correctly")
	}
	if status.ClientList[0].RealAddress != "1.2.3.4" {
		t.Errorf("Clients are not parsed correctly")
	}
	expectedClientTime, _ := time.ParseInLocation(timefmt, "Wed Apr 22 12:36:42 2020", loc)
	if !expectedClientTime.Equal(status.ClientList[0].ConnectedSince) {
		t.Errorf("Clients are not parsed correctly")
	}
}
