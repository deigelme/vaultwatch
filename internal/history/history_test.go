package history_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/history"
)

func TestTracker_NotSeenInitially(t *testing.T) {
	tr := history.New(5 * time.Minute)
	k := history.Key{SecretPath: "secret/foo", Level: "warning"}
	if tr.Seen(k) {
		t.Fatal("expected key to be unseen initially")
	}
}

func TestTracker_SeenAfterRecord(t *testing.T) {
	tr := history.New(5 * time.Minute)
	k := history.Key{SecretPath: "secret/foo", Level: "warning"}
	tr.Record(k)
	if !tr.Seen(k) {
		t.Fatal("expected key to be seen after recording")
	}
}

func TestTracker_NotSeenAfterTTLExpires(t *testing.T) {
	tr := history.New(10 * time.Millisecond)
	k := history.Key{SecretPath: "secret/bar", Level: "critical"}
	tr.Record(k)
	time.Sleep(20 * time.Millisecond)
	if tr.Seen(k) {
		t.Fatal("expected key to be unseen after TTL expires")
	}
}

func TestTracker_PurgeRemovesExpiredEntries(t *testing.T) {
	tr := history.New(10 * time.Millisecond)
	k1 := history.Key{SecretPath: "secret/a", Level: "info"}
	k2 := history.Key{SecretPath: "secret/b", Level: "warning"}
	tr.Record(k1)
	time.Sleep(20 * time.Millisecond)
	tr.Record(k2)
	tr.Purge()
	if tr.Seen(k1) {
		t.Error("k1 should have been purged")
	}
	if !tr.Seen(k2) {
		t.Error("k2 should still be present after purge")
	}
}

func TestTracker_DifferentLevelsSamePathAreIndependent(t *testing.T) {
	tr := history.New(5 * time.Minute)
	kWarn := history.Key{SecretPath: "secret/x", Level: "warning"}
	kCrit := history.Key{SecretPath: "secret/x", Level: "critical"}
	tr.Record(kWarn)
	if !tr.Seen(kWarn) {
		t.Error("warning key should be seen")
	}
	if tr.Seen(kCrit) {
		t.Error("critical key should not be seen")
	}
}
