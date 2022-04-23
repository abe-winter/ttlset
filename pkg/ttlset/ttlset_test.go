package ttlset

import "testing"
import "time"

func TestTimeComparator (t *testing.T) {
  if timeComparator(time.Time{}, time.Time{}) != 0 { t.Fatal("identical case") }
  if timeComparator(time.Time{}, time.Now()) != -1 { t.Fatal("negative case") }
  if timeComparator(time.Now(), time.Time{}) != 1 { t.Fatal("positive case") }
}

func TestGetMutex (t *testing.T) {
  if getMutex("one", keyMutexes) == nil { t.Fatal("nil mutex") }
  if getMutex("one", keyMutexes) == getMutex("two", keyMutexes) { t.Fatal("either surprising hash collision or this isn't working") }
}

// helper to check length of the Bytime value slice
func expectLen (ts *TtlSet, t *testing.T, stamp time.Time, expected int) {
  if slice, found := ts.Bytime.Get(stamp); !found {
    t.Fatal("expected found")
  } else if len(slice.([]string)) != expected {
    t.Fatalf("expected len %d got %d", expected, len(slice.([]string)))
  }
}

func TestTreeKeys (t *testing.T) {
  ts := New()
  now := time.Now()
  ts.addTreeKey("a", now)
  expectLen(&ts, t, now, 1)
  ts.addTreeKey("b", now)
  expectLen(&ts, t, now, 2)
  ts.rmTreeKey("b", now)
  expectLen(&ts, t, now, 1)
  ts.rmTreeKey("a", now)
  if _, found := ts.Bytime.Get(now); found { t.Fatal("expected not found") }
}

func TestAddRemoveLen (t *testing.T) {
  ts := New()
  now := time.Now()
  if ts.Len() != 0 { t.Fatal("wrong len") }
  if existed, _ := ts.Remove("yo", now); existed { t.Fatal("expected !existed") }
  if ts.Len() != 0 { t.Fatal("wrong len") }
  if existed, _ := ts.Add("yo", now); existed { t.Fatal("expected !existed") }
  if ts.Len() != 1 { t.Fatal("wrong len") }
  if existed, _ := ts.Remove("yo", now); !existed { t.Fatal("expected existed") }
  if ts.Len() != 0 { t.Fatal("wrong len") }
}

func TestAddExists (t *testing.T) {
  ts := New()
  base := time.Time{}
  existed, prevtime := ts.Add("yo", base)
  if existed { t.Fatal("expected !existed") }

  future := base.Add(time.Minute)
  existed, prevtime = ts.Add("yo", future)
  if !existed { t.Fatal("expected existed") }
  if !prevtime.Equal(base) { t.Fatal("wrong prevtime") }

  if !ts.Byval["yo"].t.Equal(future) { t.Fatalf("wrong time in TtlVal, wanted %s, got %s", future, ts.Byval["yo"].t) }
}

func TestCull (t *testing.T) {
  // Cull(now time.Time)
}
