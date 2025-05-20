package ids

import (
	"encoding/hex"
	"hash/fnv"
	"strings"
)

func CalculateID(longID string) string {
	h := fnv.New128()
	// We never get errors in return from FNV.
	_, _ = h.Write([]byte(longID))
	hash := hex.EncodeToString(h.Sum(nil))

	longID = normalise(longID)
	// Clamp to the first 32 characters of the long ID, then add the hash.
	if len(longID) > 32 {
		longID = longID[:32]
	}
	longID += "-" + hash[:8]
	return longID
}

func normalise(s string) string {
	var sb strings.Builder
	for _, r := range s {
		isLowercaseAscii := r >= 'a' && r <= 'z'
		isUppercaseAscii := r >= 'A' && r <= 'Z'
		isDigit := r >= '0' && r <= '9'
		if isLowercaseAscii || isUppercaseAscii || isDigit {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	return sb.String()
}
