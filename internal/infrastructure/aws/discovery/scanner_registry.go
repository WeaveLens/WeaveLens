package discovery

import "github.com/elip/WeaveLens/internal/infrastructure/aws/client"

type ScannerEntry struct {
	Name  string
	Build func(*client.Clients, string) Scanner
}

var registry []ScannerEntry

func RegisterScanner(name string, build func(*client.Clients, string) Scanner) {
	if name == "" || build == nil {
		return
	}
	registry = append(registry, ScannerEntry{Name: name, Build: build})
}

func BuildRegionScanners(clients *client.Clients, region string) []Scanner {
	if clients == nil {
		return nil
	}

	scanners := make([]Scanner, 0, len(registry))
	for _, entry := range registry {
		if scanner := entry.Build(clients, region); scanner != nil {
			scanners = append(scanners, scanner)
		}
	}
	return scanners
}
