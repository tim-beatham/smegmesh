// hosts: utility for modifying the /etc/hosts file
package hosts

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

// HOSTS_FILE is the hosts file location
const HOSTS_FILE = "/etc/hosts"

const DOMAIN_HEADER = "#WG AUTO GENERATED HOSTS"
const DOMAIN_TRAILER = "#WG AUTO GENERATED HOSTS END"

type HostsEntry struct {
	Alias string
	Ip    net.IP
}

// Generic interface to manipulate /etc/hosts file
type HostsManipulator interface {
	// AddrAddr associates an aliasd with a given IP address
	AddAddr(hosts ...HostsEntry)
	// Remove deletes the entry from /etc/hosts
	Remove(hosts ...HostsEntry)
	// Writes the changes to /etc/hosts file
	Write() error
}

type HostsManipulatorImpl struct {
	hosts map[string]HostsEntry
}

// AddAddr implements HostsManipulator.
func (m *HostsManipulatorImpl) AddAddr(hosts ...HostsEntry) {
	changed := false

	for _, host := range hosts {
		prev, ok := m.hosts[host.Ip.String()]

		if !ok || prev.Alias != host.Alias {
			changed = true
		}

		m.hosts[host.Ip.String()] = host
	}

	if changed {
		m.Write()
	}
}

// Remove implements HostsManipulator.
func (m *HostsManipulatorImpl) Remove(hosts ...HostsEntry) {
	lenBefore := len(m.hosts)

	for _, host := range hosts {
		delete(m.hosts, host.Alias)
	}

	if lenBefore != len(m.hosts) {
		m.Write()
	}
}

func (m *HostsManipulatorImpl) removeHosts() string {
	hostsFile, err := os.ReadFile(HOSTS_FILE)

	if err != nil {
		return ""
	}

	var contents strings.Builder

	scanner := bufio.NewScanner(bytes.NewReader(hostsFile))

	hostsSection := false

	for scanner.Scan() {
		line := scanner.Text()

		if err == io.EOF {
			break
		} else if err != nil {
			return ""
		}

		if !hostsSection && strings.Contains(line, DOMAIN_HEADER) {
			hostsSection = true
		}

		if !hostsSection {
			contents.WriteString(line + "\n")
		}

		if hostsSection && strings.Contains(line, DOMAIN_TRAILER) {
			hostsSection = false
		}
	}

	if scanner.Err() != nil && scanner.Err() != io.EOF {
		return ""
	}

	return contents.String()
}

// Write implements HostsManipulator
func (m *HostsManipulatorImpl) Write() error {
	contents := m.removeHosts()

	var nextHosts strings.Builder
	nextHosts.WriteString(contents)

	nextHosts.WriteString(DOMAIN_HEADER + "\n")

	for _, host := range m.hosts {
		nextHosts.WriteString(fmt.Sprintf("%s\t%s\n", host.Ip.String(), host.Alias))
	}

	nextHosts.WriteString(DOMAIN_TRAILER + "\n")
	return os.WriteFile(HOSTS_FILE, []byte(nextHosts.String()), 0644)
}

func NewHostsManipulator() HostsManipulator {
	return &HostsManipulatorImpl{hosts: make(map[string]HostsEntry)}
}
