package store


import (
	"sync"
	"time"
)


type BlockEvent struct {
	Hash string    `json:"hash"`
	Ts   int64     `json:"ts"`
	Time time.Time `json:"-"`
}


type TxEvent struct {
	Txid string    `json:"txid"`
	Ts   int64     `json:"ts"`
	Time time.Time `json:"-"`
}


type EventStore struct {
	mu sync.RWMutex

	blocks    []BlockEvent
	txs       []TxEvent
	maxBlocks int
	maxTxs    int

	
	startTime     time.Time
	totalTxCount  int64
	totalBlkCount int64
}


func New(maxBlocks, maxTxs int) *EventStore {
	return &EventStore{
		blocks:    make([]BlockEvent, 0, maxBlocks),
		txs:       make([]TxEvent, 0, maxTxs),
		maxBlocks: maxBlocks,
		maxTxs:    maxTxs,
		startTime: time.Now(),
	}
}

func (s *EventStore) AddBlock(hash string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	ev := BlockEvent{Hash: hash, Ts: now.Unix(), Time: now}

	if len(s.blocks) >= s.maxBlocks {
		
		s.blocks = s.blocks[1:]
	}
	s.blocks = append(s.blocks, ev)
	s.totalBlkCount++
}


func (s *EventStore) AddTx(txid string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	ev := TxEvent{Txid: txid, Ts: now.Unix(), Time: now}

	if len(s.txs) >= s.maxTxs {
		s.txs = s.txs[1:]
	}
	s.txs = append(s.txs, ev)
	s.totalTxCount++
}


type Summary struct {
	BlocksObserved int64   `json:"blocks_observed"`
	TxObserved     int64   `json:"tx_observed"`
	LastEventTime  int64   `json:"last_event_time"`
	TxPerSecond    float64 `json:"tx_per_second"`
	UptimeSeconds  float64 `json:"uptime_seconds"`
}

func (s *EventStore) GetSummary() Summary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	elapsed := time.Since(s.startTime).Seconds()
	txPerSec := 0.0
	if elapsed > 0 {
		txPerSec = float64(s.totalTxCount) / elapsed
	}

	lastTs := int64(0)
	if len(s.blocks) > 0 {
		lastTs = s.blocks[len(s.blocks)-1].Ts
	}
	if len(s.txs) > 0 && s.txs[len(s.txs)-1].Ts > lastTs {
		lastTs = s.txs[len(s.txs)-1].Ts
	}

	return Summary{
		BlocksObserved: s.totalBlkCount,
		TxObserved:     s.totalTxCount,
		LastEventTime:  lastTs,
		TxPerSecond:    round2(txPerSec),
		UptimeSeconds:  round2(elapsed),
	}
}


type Latest struct {
	Blocks []BlockEvent `json:"blocks"`
	Txs    []TxEvent    `json:"txs"`
}

func (s *EventStore) GetLatest(nBlocks, nTxs int) Latest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blocks := make([]BlockEvent, 0)
	if len(s.blocks) > 0 {
		start := len(s.blocks) - nBlocks
		if start < 0 {
			start = 0
		}
		
		for i := len(s.blocks) - 1; i >= start; i-- {
			blocks = append(blocks, s.blocks[i])
		}
	}

	txs := make([]TxEvent, 0)
	if len(s.txs) > 0 {
		start := len(s.txs) - nTxs
		if start < 0 {
			start = 0
		}
		for i := len(s.txs) - 1; i >= start; i-- {
			txs = append(txs, s.txs[i])
		}
	}

	return Latest{Blocks: blocks, Txs: txs}
}


func (s *EventStore) LastSeenBlock() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.blocks) == 0 {
		return ""
	}
	return s.blocks[len(s.blocks)-1].Hash
}

func round2(v float64) float64 {
	return float64(int(v*100)) / 100
}