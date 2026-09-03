package transport

import (
	"encoding/json"
	"sync"
)

type ScanBroadcaster struct {
	mu          sync.RWMutex
	subscribers map[chan []byte]struct{}
}

func NewScanBroadcaster() *ScanBroadcaster {
	return &ScanBroadcaster{
		subscribers: make(map[chan []byte]struct{}),
	}
}

func (b *ScanBroadcaster) Subscribe() chan []byte {
	ch := make(chan []byte, 16)
	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *ScanBroadcaster) Unsubscribe(ch chan []byte) {
	b.mu.Lock()
	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}
	b.mu.Unlock()
}

func (b *ScanBroadcaster) Broadcast(scans any) {
	data, err := json.Marshal(scans)
	if err != nil {
		return
	}
	payload := append([]byte("data: "), data...)
	payload = append(payload, '\n', '\n')

	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subscribers {
		select {
		case ch <- payload:
		default:
		}
	}
}
