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

const badFields = `OpenVPN CLIENT LIST
Updated,test
Common Name,Real Address,Bytes Received,Bytes Sent,Connected Since
user1,1.2.3.4,foo,foo,test
ROUTING TABLE
Virtual Address,Common Name,Real Address,Last Ref
10.240.1.222,user4,1.2.3.7:fooo,test
GLOBAL STATS
Max bcast/mcast queue length,foo
END
`

func TestParsingWrongValuesIsNotAnIssue(t *testing.T) {
	status, e := parse(bufio.NewReader(strings.NewReader(badFields)))
	if e != nil {
		t.Errorf("should have worked")
	}
	if status.GlobalStats.MaxBcastMcastQueueLen != 0 {
		t.Errorf("Parsing wrong MaxBcastMcastQueueLen value lead to unexpected result")
	}
	expectedTime := time.Time{}
	if !expectedTime.Equal(status.UpdatedAt) {
		t.Errorf("parsing incorrect time value should have yieleded a default time object")
	}
}

const connectedClientsV2 = `TITLE,OpenVPN 2.4.4 x86_64-pc-linux-gnu [SSL (OpenSSL)] [LZO] [LZ4] [EPOLL] [PKCS11] [MH/PKTINFO] [AEAD] built on May 14 2019
TIME,Thu Apr 30 13:55:44 2020,1588254944
HEADER,CLIENT_LIST,Common Name,Real Address,Virtual Address,Virtual IPv6 Address,Bytes Received,Bytes Sent,Connected Since,Connected Since (time_t),Username,Client ID,Peer ID
CLIENT_LIST,test@localhost,1.2.3.4:54190,10.80.0.65,,3860,3688,Thu Apr 30 13:55:38 2020,1588254938,test@localhost,0,0
CLIENT_LIST,test1@localhost,1.2.3.5:51053,10.68.0.25,,3871,3924,Thu Apr 30 13:55:40 2020,1588254940,test1@localhost,1,1
HEADER,ROUTING_TABLE,Virtual Address,Common Name,Real Address,Last Ref,Last Ref (time_t)
ROUTING_TABLE,10.80.0.65,test@localhost,1.2.3.4:54190,Thu Apr 30 13:55:40 2020,1588254940
ROUTING_TABLE,10.68.0.25,test1@localhost,1.2.3.5:51053,Thu Apr 30 13:55:42 2020,1588254942
GLOBAL_STATS,Max bcast/mcast queue length,0
END
`
func TestConnectedClientsParsedCorrectlyWithStatusVersion2(t *testing.T) {
	_, e := parse(bufio.NewReader(strings.NewReader(connectedClientsV2)))
	if e != nil {
		t.Errorf("should have worked")
	}
}
