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

// Generic interface to manipulate /etc/hosts file
type HostsManipulator interface {
	// AddrAddr associates an aliasd with a given IP address
	AddAddr(ipAddr net.IP, alias string)
	// Remove deletes the entry from /etc/hosts
	Remove(alias string)
	// Writes the changes to /etc/hosts file
	Write() error
}

type HostsManipulatorImpl struct {
	hosts  map[string]net.IP
	meshid string
}

// AddAddr implements HostsManipulator.
func (m *HostsManipulatorImpl) AddAddr(ipAddr net.IP, alias string) {
	m.hosts[alias] = ipAddr
}

// Remove implements HostsManipulator.
func (m *HostsManipulatorImpl) Remove(alias string) {
	delete(m.hosts, alias)
}

type HostsEntry struct {
	Alias string
	Ip    net.IP
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

		if !hostsSection && strings.Contains(line, DOMAIN_HEADER+m.meshid) {
			hostsSection = true
		}

		if !hostsSection {
			contents.WriteString(line + "\n")
		}

		if hostsSection && strings.Contains(line, DOMAIN_TRAILER+m.meshid) {
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

	nextHosts.WriteString(DOMAIN_HEADER + m.meshid + "\n")

	for alias, ip := range m.hosts {
		nextHosts.WriteString(fmt.Sprintf("%s\t%s\n", ip.String(), alias))
	}

	nextHosts.WriteString(DOMAIN_TRAILER + m.meshid + "\n")
	return os.WriteFile(HOSTS_FILE, []byte(nextHosts.String()), 0644)
}

// parseLine parses a line in the /etc/hosts file
func parseLine(line string) (*HostsEntry, error) {
	fields := strings.Fields(line)

	if len(fields) != 2 {
		return nil, fmt.Errorf("expected entry length of 2 was %d", len(fields))
	}

	ipAddr := fields[0]
	alias := fields[1]

	ip := net.ParseIP(ipAddr)

	if ip == nil {
		return nil, fmt.Errorf("failed to parse ip for %s", alias)
	}

	return &HostsEntry{Ip: ip, Alias: alias}, nil
}

func NewHostsManipulator(meshId string) HostsManipulator {
	return &HostsManipulatorImpl{hosts: make(map[string]net.IP), meshid: meshId}
}
