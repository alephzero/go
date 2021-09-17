package alephzero

import (
	"os"
	"testing"
)

func TestFile(t *testing.T) {
	file, err := FileOpen("foo", nil)
	check(t, err)
	path := file.Path()
	if path != "/dev/shm/foo" {
		t.Errorf("file.Path() = %q, want foo", path)
	}
	fd := file.Fd()
	if fd <= 2 {
		t.Errorf("file.Fd() = %d, want > 2", fd)
	}
	arena := file.Arena()
	if len(arena.Buf()) != 16*1024*1024 {
		t.Errorf("len(file.Arena().Buf()) = %d, want 16MB", len(arena.Buf()))
	}
	check(t, file.Close())

	if _, err := os.Stat("/dev/shm/foo"); os.IsNotExist(err) {
		t.Errorf("/dev/shm/foo should exist")
	}

	check(t, FileRemove("foo"))

	if _, err := os.Stat("/dev/shm/foo"); !os.IsNotExist(err) {
		t.Errorf("/dev/shm/foo should not exist")
	}

	opts := MakeDefaultFileOptions()
	opts.CreateOptions.Size = 1024

	if opts.OpenOptions.ArenaMode != MODE_SHARED {
		t.Errorf("opts.OpenOptions.ArenaMode = %d, want %d", opts.OpenOptions.ArenaMode, MODE_SHARED)
	}

	file, err = FileOpen("foo", &opts)
	check(t, err)
	path = file.Path()
	if path != "/dev/shm/foo" {
		t.Errorf("file.Path() = %q, want foo", path)
	}
	fd = file.Fd()
	if fd <= 2 {
		t.Errorf("file.Fd() = %d, want > 2", fd)
	}
	arena = file.Arena()
	if len(arena.Buf()) != 1024 {
		t.Errorf("len(file.Arena().Buf()) = %d, want 1kB", len(arena.Buf()))
	}
	check(t, file.Close())

	if _, err = os.Stat("/dev/shm/foo"); os.IsNotExist(err) {
		t.Errorf("/dev/shm/foo should exist")
	}

	check(t, FileRemove("foo"))

	if _, err = os.Stat("/dev/shm/foo"); !os.IsNotExist(err) {
		t.Errorf("/dev/shm/foo should not exist")
	}
}
