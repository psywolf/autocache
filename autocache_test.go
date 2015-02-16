package autocache

import (
	"testing"
)

var words = map[string]string{
	"et":        "et",
	"oratio":    "oratio",
	"conviciis": "convicium",
	"est":       "sum",
}

func TestGet(t *testing.T) {
	hitCount := 0
	c := New(3, func(key string) (string, error) {
		hitCount++
		return words[key], nil
	})

	for _, word := range []string{"et", "oratio", "et", "conviciis", "et", "est", "oratio"} {
		val, _ := c.Get(word)
		t.Log(val)
		t.Log(hitCount)
	}

	expectedCount := 5
	if hitCount != expectedCount {
		t.Errorf("Hitcount was %d.  Expected hitcount of %d", hitCount, expectedCount)
	}
}
