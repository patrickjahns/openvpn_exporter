package openvpn

import (
	"math/rand"
	"strings"
	"time"
)

type PseudonymizingDecorator struct {
	pseudonymizeClientMetricsLength     int
	pseudonymizeClientMetricsMap        map[string]string
	reversePseudonymizeClientMetricsMap map[string]string
}

func NewOpenVPNPseudonymizingDecorator(
	pseudonymizeClientMetricsLength int,
) PseudonymizingDecorator {
	return PseudonymizingDecorator{
		pseudonymizeClientMetricsLength:     pseudonymizeClientMetricsLength,
		pseudonymizeClientMetricsMap:        make(map[string]string),
		reversePseudonymizeClientMetricsMap: make(map[string]string),
	}
}

func reverseMap(m map[string]string) map[string]string {
	n := make(map[string]string, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}

// see https://stackoverflow.com/a/22892986/18529703
func (t PseudonymizingDecorator) generatePseudonym(n int) string {
	src := rand.NewSource(time.Now().UnixNano())
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6
		letterIdxMask = 1<<letterIdxBits - 1
		letterIdxMax  = 63 / letterIdxBits
	)

	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func (t PseudonymizingDecorator) DecorateParseFile(f func(statusfile string) (*Status, error)) func(statusfile string) (*Status, error) {
	return func(statusfile string) (*Status, error) {
		status, err := f(statusfile)
		if err != nil {
			return nil, err
		}

		// use index for iteration to work on the actual reference instead of a value
		for idx := range status.ClientList {
			client := &status.ClientList[idx]
			commonName := client.CommonName
			pseudonym, ok := t.pseudonymizeClientMetricsMap[commonName]
			if !ok {
				for {
					pseudonym = t.generatePseudonym(t.pseudonymizeClientMetricsLength)
					if _, ok := t.reversePseudonymizeClientMetricsMap[pseudonym]; !ok {
						break
					}
				}
			}

			t.pseudonymizeClientMetricsMap[commonName] = pseudonym
			// update the reverse map for quick lookup of existing pseudonyms
			t.reversePseudonymizeClientMetricsMap = reverseMap(t.pseudonymizeClientMetricsMap)

			client.CommonName = pseudonym
		}

		return status, nil
	}
}
