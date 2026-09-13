package bob

import "testing"

func TestAcquireThenReleaseAllowsReacquire(t *testing.T) {
	dir := t.TempDir()
	release, err := Acquire(dir)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	release()
	if _, err := Acquire(dir); err != nil {
		t.Fatalf("Acquire after release: %v", err)
	}
}

func TestAcquireRejectsReentrantLock(t *testing.T) {
	dir := t.TempDir()
	release, err := Acquire(dir)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	defer release()

	if _, err := Acquire(dir); err == nil {
		t.Fatalf("second Acquire() = nil error while lock held, want refusal")
	}
}
