package eventstore

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"heritage-tree-relocation-clearance/internal/domain"
)

const schemaVersion = 1

type EventRecord struct {
	Sequence    uint64          `json:"sequence"`
	EventIndex  int             `json:"eventIndex"`
	CaseVersion uint64          `json:"caseVersion"`
	Type        string          `json:"type"`
	CaseID      string          `json:"caseID"`
	OccurredAt  time.Time       `json:"occurredAt"`
	Actor       string          `json:"actor"`
	Data        json.RawMessage `json:"data"`
}

type frame struct {
	SchemaVersion   int            `json:"schemaVersion"`
	Sequence        uint64         `json:"sequence"`
	PreviousDigest  string         `json:"previousDigest"`
	CaseID          string         `json:"caseID"`
	ExpectedVersion uint64         `json:"expectedVersion"`
	IdempotencyKey  string         `json:"idempotencyKey"`
	Events          []domain.Event `json:"events"`
	Checksum        string         `json:"checksum"`
}

type idempotencyResult struct {
	CaseID   string            `json:"caseID"`
	Version  uint64            `json:"version"`
	Sequence uint64            `json:"sequence"`
	Status   domain.CaseStatus `json:"status"`
}

type snapshot struct {
	SchemaVersion int                               `json:"schemaVersion"`
	Sequence      uint64                            `json:"sequence"`
	LogDigest     string                            `json:"logDigest"`
	NextSerial    uint64                            `json:"nextSerial"`
	Cases         map[string]*domain.RelocationCase `json:"cases"`
	Idempotency   map[string]idempotencyResult      `json:"idempotency"`
	Checksum      string                            `json:"checksum"`
}

type Store struct {
	mu           sync.RWMutex
	dir          string
	logPath      string
	snapshotPath string
	lockFile     *os.File
	sequence     uint64
	lastDigest   string
	nextSerial   uint64
	cases        map[string]*domain.RelocationCase
	events       map[string][]EventRecord
	credentials  map[string]domain.ClearanceCredential
	idempotency  map[string]idempotencyResult
	closed       bool
}

type AppendResult struct {
	Version    uint64            `json:"version"`
	Sequence   uint64            `json:"sequence"`
	Status     domain.CaseStatus `json:"status"`
	Idempotent bool              `json:"idempotent"`
}
