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

const noConnectedClientsV1 = `OpenVPN CLIENT LIST
Updated,Thu Apr 23 20:14:31 2020
Common Name,Real Address,Bytes Received,Bytes Sent,Connected Since
ROUTING TABLE
Virtual Address,Common Name,Real Address,Last Ref
GLOBAL STATS
Max bcast/mcast queue length,0
END
`
const noConnectedClientsV2 = `TITLE,OpenVPN 2.4.4 x86_64-pc-linux-gnu [SSL (OpenSSL)] [LZO] [LZ4] [EPOLL] [PKCS11] [MH/PKTINFO] [AEAD] built on May 14 2019
TIME,Thu Apr 30 13:55:44 2020,1588254944
HEADER,CLIENT_LIST,Common Name,Real Address,Virtual Address,Virtual IPv6 Address,Bytes Received,Bytes Sent,Connected Since,Connected Since (time_t),Username,Client ID,Peer ID
HEADER,ROUTING_TABLE,Virtual Address,Common Name,Real Address,Last Ref,Last Ref (time_t)
GLOBAL_STATS,Max bcast/mcast queue length,0
END
`
const noConnectedClientsV3 = `TITLE	OpenVPN 2.4.4 x86_64-pc-linux-gnu [SSL (OpenSSL)] [LZO] [LZ4] [EPOLL] [PKCS11] [MH/PKTINFO] [AEAD] built on May 14 2019
TIME	Thu Apr 30 13:55:44 2020	1588254944
HEADER	CLIENT_LIST	Common Name	Real Address	Virtual Address	Virtual IPv6 Address	Bytes Received	Bytes Sent	Connected Since	Connected Since (time_t)	Username	Client ID	Peer ID
HEADER	ROUTING_TABLE	Virtual Address	Common Name	Real Address	Last Ref	Last Ref (time_t)
GLOBAL_STATS	Max bcast/mcast queue length	0
END
`

var noConnectedClientsTestCases = []struct {
	StatusVersionName  string
	StatusFileContents string
}{
	{"v1", noConnectedClientsV1},
	{"v2", noConnectedClientsV2},
	{"v3", noConnectedClientsV3},
}

func TestNoConnectedClientsAreParsedCorrectly(t *testing.T) {
	for _, tt := range noConnectedClientsTestCases {
		t.Run(tt.StatusVersionName, func(t *testing.T) {
			status, e := parse(bufio.NewReader(strings.NewReader(tt.StatusFileContents)))
			if e != nil {
				t.Errorf("should have worked")
			}
			if len(status.ClientList) != 0 {
				t.Errorf("Clients are not parsed correctly")
			}
			if status.GlobalStats.MaxBcastMcastQueueLen != 0 {
				t.Errorf("MaxBcastMcastQueueLen was not parsed correctly")
			}
		})
	}
}

const connectedClientsV1 = `OpenVPN CLIENT LIST
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
const connectedClientsV3 = `TITLE	OpenVPN 2.4.4 x86_64-pc-linux-gnu [SSL (OpenSSL)] [LZO] [LZ4] [EPOLL] [PKCS11] [MH/PKTINFO] [AEAD] built on May 14 2019
TIME	Thu Apr 30 13:55:44 2020	1588254944
HEADER	CLIENT_LIST	Common Name	Real Address	Virtual Address	Virtual IPv6 Address	Bytes Received	Bytes Sent	Connected Since	Connected Since (time_t)	Username	Client ID	Peer ID
CLIENT_LIST	test@localhost	1.2.3.4:54190	10.80.0.65		3860	3688	Thu Apr 30 13:55:38 2020	1588254938	test@localhost	0	0
CLIENT_LIST	test1@localhost	1.2.3.5:51053	10.68.0.25		3871	3924	Thu Apr 30 13:55:40 2020	1588254940	test1@localhost	1	1
HEADER	ROUTING_TABLE	Virtual Address	Common Name	Real Address	Last Ref	Last Ref (time_t)
ROUTING_TABLE	10.80.0.65	test@localhost	1.2.3.4:54190	Thu Apr 30 13:55:40 2020	1588254940
ROUTING_TABLE	10.68.0.25	test1@localhost	1.2.3.5:51053	Thu Apr 30 13:55:42 2020	1588254942
GLOBAL_STATS	Max bcast/mcast queue length	0
END
`

func parseDate(dateToParse string) time.Time {
	loc, _ := time.LoadLocation("Local")
	expectedTime, _ := time.ParseInLocation(timefmt, dateToParse, loc)
	return expectedTime
}

var correctlyParsedTestCases = []struct {
	StatusVersionName        string
	StatusFileContents       string
	UpdatedAt                time.Time
	NumberOfConnectedClients int
	MaxBcastMcastQueueLength int
	Client0CommonNamme       string
	Client0Address           string
	Client0ConnectedSince    time.Time
}{
	{"v1", connectedClientsV1, parseDate("Thu Apr 23 20:14:31 2020"), 4, 5, "user1", "1.2.3.4", parseDate("Wed Apr 22 12:36:42 2020")},
	{"v2", connectedClientsV2, time.Unix(1588254944, 0), 2, 0, "test@localhost", "1.2.3.4", time.Unix(1588254938, 0)},
	{"v3", connectedClientsV3, time.Unix(1588254944, 0), 2, 0, "test@localhost", "1.2.3.4", time.Unix(1588254938, 0)},
}

func TestConnectedClientsParsedCorrectly(t *testing.T) {
	for _, tt := range correctlyParsedTestCases {
		t.Run(tt.StatusVersionName, func(t *testing.T) {
			status, e := parse(bufio.NewReader(strings.NewReader(tt.StatusFileContents)))
			if e != nil {
				t.Errorf("should have worked")
			}
			if !tt.UpdatedAt.Equal(status.UpdatedAt) {
				t.Errorf("failed parsing updated at")
			}
			if tt.MaxBcastMcastQueueLength != status.GlobalStats.MaxBcastMcastQueueLen {
				t.Errorf("failed parsing bcast/mcast queue length")
			}
			if len(status.ClientList) != tt.NumberOfConnectedClients {
				t.Errorf("Clients are not parsed correctly")
			}
			if status.ClientList[0].CommonName != tt.Client0CommonNamme {
				t.Errorf("Clients are not parsed correctly")
			}
			if status.ClientList[0].RealAddress != tt.Client0Address {
				t.Errorf("Clients are not parsed correctly")
			}
			if !tt.Client0ConnectedSince.Equal(status.ClientList[0].ConnectedSince) {
				t.Errorf("Clients are not parsed correctly")
			}
		})
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

var serverInfoTestCases = []struct {
	StatusVersionName        string
	StatusFileContents       string
	ServerInfoVersion        string
	ServerInfoArch           string
	ServerInfoAdditionalInfo string
}{
	{"v1", connectedClientsV1, "unknown", "unknown", "unknown"},
	{"v2", connectedClientsV2, "2.4.4", "x86_64-pc-linux-gnu", "[SSL (OpenSSL)] [LZO] [LZ4] [EPOLL] [PKCS11] [MH/PKTINFO] [AEAD] built on May 14 2019"},
	{"v3", connectedClientsV3, "2.4.4", "x86_64-pc-linux-gnu", "[SSL (OpenSSL)] [LZO] [LZ4] [EPOLL] [PKCS11] [MH/PKTINFO] [AEAD] built on May 14 2019"},
}

func TestServerInfoIsParsedCorrectly(t *testing.T) {
	for _, tt := range serverInfoTestCases {
		t.Run(tt.StatusVersionName, func(t *testing.T) {
			status, _ := parse(bufio.NewReader(strings.NewReader(tt.StatusFileContents)))
			if status.ServerInfo.Version != tt.ServerInfoVersion {
				t.Errorf("version is not parsed correctly")
			}
			if status.ServerInfo.Arch != tt.ServerInfoArch {
				t.Errorf("arch is not parsed correctly")
			}
			if status.ServerInfo.AdditionalInfo != tt.ServerInfoAdditionalInfo {
				t.Errorf("additional info is not parsed correctly")
			}
		})
	}

}
