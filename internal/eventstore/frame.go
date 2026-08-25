package eventstore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func frameChecksum(f frame) (string, error) {
	f.Checksum = ""
	b, err := json.Marshal(f)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func frameDigest(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func snapshotChecksum(s snapshot) (string, error) {
	s.Checksum = ""
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
