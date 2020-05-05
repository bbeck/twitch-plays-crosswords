package main

import (
	"sync"
)

// ChannelMonitor gets notified about current puzzles being solved on a per
// channel basis and determines which channels have been added or removed and
// notifies callbacks of the changes.
type ChannelMonitor struct {
	sync.Mutex

	// the set of channels being monitored on a per-integration basis
	channels map[ID]map[string]struct{}

	// callback to call when a channel has been added to the monitored list
	OnAddChannel func(channel string)

	// callback to call when a channel has been removed from the monitored list
	OnRemoveChannel func(channel string)

	// callback to call whenever an update of channels is received
	OnUpdateChannels func(channels map[ID][]string)
}

// Update records the updated set of channels from a particular integration's
// channel locator.
func (m *ChannelMonitor) Update(update map[ID][]string) {
	// Lock so that we have a consistent picture of the world.
	m.Lock()
	defer m.Unlock()

	if m.channels == nil {
		m.channels = make(map[ID]map[string]struct{})
	}

	// Compute the set of channels that we were monitoring before this update.
	before := m.AllChannels()

	// Apply the update.
	for id, channels := range update {
		m.channels[id] = make(map[string]struct{})

		for _, channel := range channels {
			m.channels[id][channel] = struct{}{}
		}
	}

	// Now compute the set of channels we are monitoring after this update.
	after := m.AllChannels()

	// Call the global channel callbacks.
	added, removed := m.Diff(before, after)
	for _, channel := range added {
		if m.OnAddChannel != nil {
			m.OnAddChannel(channel)
		}
	}
	for _, channel := range removed {
		if m.OnRemoveChannel != nil {
			m.OnRemoveChannel(channel)
		}
	}

	// Call the integration callback.
	if m.OnUpdateChannels != nil {
		m.OnUpdateChannels(update)
	}
}

// Diff determines which channels are new and which are removed.
func (m *ChannelMonitor) Diff(before, after map[string]bool) ([]string, []string) {
	all := make(map[string]struct{})
	for channel := range before {
		all[channel] = struct{}{}
	}
	for channel := range after {
		all[channel] = struct{}{}
	}

	var added, removed []string
	for channel := range all {
		if !before[channel] && after[channel] {
			added = append(added, channel)
		}
		if before[channel] && !after[channel] {
			removed = append(removed, channel)
		}
	}

	return added, removed
}

// AllChannels calculates the union of all channels being monitored.
func (m *ChannelMonitor) AllChannels() map[string]bool {
	seen := make(map[string]bool)
	for _, cs := range m.channels {
		for c := range cs {
			seen[c] = true
		}
	}

	return seen
}