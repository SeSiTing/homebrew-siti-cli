package cmd

import (
	"bytes"
	"testing"

	"github.com/SeSiTing/siti-cli/internal/network"
)

func TestPrintNetworkTransition(t *testing.T) {
	before := network.LiveStatus{
		Mode:       "DHCP Configuration",
		Address:    "172.16.40.69",
		SubnetMask: "255.255.240.0",
		Gateway:    "172.16.32.1",
	}
	after := network.LiveStatus{
		Mode:       "Manual Configuration",
		Address:    "172.16.40.69",
		SubnetMask: "255.255.240.0",
		Gateway:    "172.16.40.2",
		DNS:        []string{"172.16.40.2"},
	}

	var output bytes.Buffer
	printNetworkTransition(&output, before, after)
	want := `  changes:
    mode: DHCP Configuration -> Manual Configuration
    address: 172.16.40.69 -> 172.16.40.69
    subnet mask: 255.255.240.0 -> 255.255.240.0
    gateway: 172.16.32.1 -> 172.16.40.2
    DNS: Automatic -> 172.16.40.2
`
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}
