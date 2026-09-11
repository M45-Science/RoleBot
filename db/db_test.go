package db

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteClusterWritesCompactCompatibleRecord(t *testing.T) {
	workdir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previousDir) })

	if err := os.MkdirAll(filepath.Join("data", "db"), 0755); err != nil {
		t.Fatal(err)
	}

	previousDatabase := Database
	Database = [16384]*GuildData{}
	t.Cleanup(func() { Database = previousDatabase })

	Database[1] = &GuildData{
		LID:     1,
		Guild:   42,
		Added:   3,
		Donator: 1,
		Roles:   []RoleData{{ID: 100}, {ID: 200}},
	}
	WriteCluster(0)

	data, err := os.ReadFile(filepath.Join("data", "db", "cluster-1.dat"))
	if err != nil {
		t.Fatal(err)
	}

	wantSize := VERSION_size + staticSize + 2*SNOWFLAKE_size
	if len(data) != wantSize {
		t.Fatalf("cluster size = %d, want %d", len(data), wantSize)
	}
	if version := binary.LittleEndian.Uint16(data); version != 1 {
		t.Fatalf("version = %d, want 1", version)
	}
	if guild := binary.LittleEndian.Uint64(data[VERSION_size+LID_size:]); guild != 42 {
		t.Fatalf("guild = %d, want 42", guild)
	}
}
