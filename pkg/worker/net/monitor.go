package net

import (
	"context"
	stdnet "net"
	"sort"
	"strings"
	"time"
)

const defaultPollInterval = 5 * time.Second

// NetMonitor polls network interfaces for address changes and
// triggers validation when a change is detected.
type NetMonitor struct {
	interval   time.Duration
	onChange   func(oldAddrs, newAddrs []string)
	onValidate func() error
}

// NewNetMonitor creates a monitor that polls every 5s.
// When addresses change and onValidate returns an error, onChange is called.
func NewNetMonitor(onChange func([]string, []string), onValidate func() error) *NetMonitor {
	return &NetMonitor{
		interval:   defaultPollInterval,
		onChange:   onChange,
		onValidate: onValidate,
	}
}

// Start runs the polling loop in a goroutine until ctx is cancelled.
func (m *NetMonitor) Start(ctx context.Context) {
	go m.run(ctx)
}

func (m *NetMonitor) run(ctx context.Context) {
	prev := currentAddrs()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cur := currentAddrs()
			if !equal(prev, cur) {
				if err := m.onValidate(); err != nil {
					m.onChange(prev, cur)
				}
				prev = cur
			}
		}
	}
}

// currentAddrs returns a sorted list of all unicast addresses.
func currentAddrs() []string {
	ifaces, err := stdnet.Interfaces()
	if err != nil {
		return nil
	}

	var addrs []string
	for _, iface := range ifaces {
		ifAddrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range ifAddrs {
			addrs = append(addrs, a.String())
		}
	}
	sort.Strings(addrs)
	return addrs
}

// equal compares two sorted string slices.
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	return strings.Join(a, ",") == strings.Join(b, ",")
}
