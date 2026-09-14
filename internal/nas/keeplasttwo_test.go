package nas

import (
	"fmt"
	"testing"
)

func withMockedExec(t *testing.T, list func() ([]byte, error), del func(path string) ([]byte, error)) {
	t.Helper()
	origList, origDel := execListBackupFolders, execDeleteFolder
	if list != nil {
		execListBackupFolders = list
	}
	if del != nil {
		execDeleteFolder = del
	}
	t.Cleanup(func() {
		execListBackupFolders = origList
		execDeleteFolder = origDel
	})
}

func TestKeepLastTwo_ListFails(t *testing.T) {
	withMockedExec(t, func() ([]byte, error) {
		return nil, fmt.Errorf("ssh: connection refused")
	}, nil)

	if err := KeepLastTwo(); err == nil {
		t.Fatal("expected an error when listing fails, got nil")
	}
}

func TestKeepLastTwo_FewerThanThreeFolders_NoDeletion(t *testing.T) {
	var deleteCalled bool
	withMockedExec(t, func() ([]byte, error) {
		return []byte("drwxr-xr-x 2 root root 4096 2024-01-01\ndrwxr-xr-x 2 root root 4096 2024-01-02\n"), nil
	}, func(path string) ([]byte, error) {
		deleteCalled = true
		return nil, nil
	})

	if err := KeepLastTwo(); err != nil {
		t.Fatalf("KeepLastTwo() returned unexpected error: %v", err)
	}
	if deleteCalled {
		t.Error("expected no deletion with fewer than 3 folders, but deletion was called")
	}
}

func TestKeepLastTwo_DeletesOldestFolder(t *testing.T) {
	var deletedPath string
	withMockedExec(t, func() ([]byte, error) {
		return []byte(
			"drwxr-xr-x 2 root root 4096 2024-08-28\n" +
				"drwxr-xr-x 2 root root 4096 2024-09-04\n" +
				"drwxr-xr-x 2 root root 4096 2024-09-11\n",
		), nil
	}, func(path string) ([]byte, error) {
		deletedPath = path
		return nil, nil
	})

	if err := KeepLastTwo(); err != nil {
		t.Fatalf("KeepLastTwo() returned unexpected error: %v", err)
	}
	want := "/mnt/datastore/backupExternal/2024-08-28"
	if deletedPath != want {
		t.Errorf("deleted path = %q, want %q (the oldest folder)", deletedPath, want)
	}
}

func TestKeepLastTwo_DeleteFails(t *testing.T) {
	withMockedExec(t, func() ([]byte, error) {
		return []byte("2024-08-28\n2024-09-04\n2024-09-11\n"), nil
	}, func(path string) ([]byte, error) {
		return nil, fmt.Errorf("permission denied")
	})

	if err := KeepLastTwo(); err == nil {
		t.Fatal("expected an error when deletion fails, got nil")
	}
}
