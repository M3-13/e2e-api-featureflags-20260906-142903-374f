package evaluate

import "hash/fnv"

func Decide(key, userID string, rolloutPercent int, enabled bool) bool {
	if !enabled {
		return false
	}
	if rolloutPercent <= 0 {
		return false
	}
	if rolloutPercent >= 100 {
		return true
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(key + "\x00" + userID))
	return int(h.Sum32()%100) < rolloutPercent
}
