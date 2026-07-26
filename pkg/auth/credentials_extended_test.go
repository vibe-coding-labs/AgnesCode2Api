package auth

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestLoadFromSystem_NonDarwin(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("skipping non-darwin test on darwin")
	}
	_, err := LoadFromSystem()
	if err == nil {
		t.Fatal("expected error on non-darwin platform, got nil")
	}
}

func TestCredentials_EmptyFields(t *testing.T) {
	creds := &Credentials{}
	if creds.JWTToken != "" {
		t.Errorf("JWTToken = %q, want empty string", creds.JWTToken)
	}
	if creds.UserID != "" {
		t.Errorf("UserID = %q, want empty string", creds.UserID)
	}
}

func TestCredentials_NilVsNonNil(t *testing.T) {
	var creds *Credentials
	nonNil := &Credentials{JWTToken: "x", UserID: "y"}
	if creds == nonNil {
		t.Error("nil should not equal allocated Credentials")
	}
}

func TestStateData_JSONParsing(t *testing.T) {
	raw := `{"agnesCoderUser":{"JWTToken":"test-jwt-token-123","userId":"user-456"}}`
	var data stateData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatalf("failed to parse valid JSON: %v", err)
	}
	if data.AgnesCoderUser.JWTToken != "test-jwt-token-123" {
		t.Errorf("JWTToken = %q, want %q", data.AgnesCoderUser.JWTToken, "test-jwt-token-123")
	}
	if data.AgnesCoderUser.UserID != "user-456" {
		t.Errorf("UserID = %q, want %q", data.AgnesCoderUser.UserID, "user-456")
	}
}

func TestStateData_EmptyJWTToken(t *testing.T) {
	raw := `{"agnesCoderUser":{"JWTToken":"","userId":"user-789"}}`
	var data stateData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if data.AgnesCoderUser.JWTToken != "" {
		t.Errorf("JWTToken = %q, want empty string", data.AgnesCoderUser.JWTToken)
	}
	if data.AgnesCoderUser.UserID != "user-789" {
		t.Errorf("UserID = %q, want %q", data.AgnesCoderUser.UserID, "user-789")
	}
}

func TestStateData_EmptyUserID(t *testing.T) {
	raw := `{"agnesCoderUser":{"JWTToken":"some-key","userId":""}}`
	var data stateData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if data.AgnesCoderUser.JWTToken != "some-key" {
		t.Errorf("JWTToken = %q, want %q", data.AgnesCoderUser.JWTToken, "some-key")
	}
	if data.AgnesCoderUser.UserID != "" {
		t.Errorf("UserID = %q, want empty string", data.AgnesCoderUser.UserID)
	}
}

func TestStateData_MissingAgnesCoderUser(t *testing.T) {
	raw := `{"otherField":"some-value"}`
	var data stateData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if data.AgnesCoderUser.JWTToken != "" {
		t.Errorf("JWTToken = %q, want empty string", data.AgnesCoderUser.JWTToken)
	}
	if data.AgnesCoderUser.UserID != "" {
		t.Errorf("UserID = %q, want empty string", data.AgnesCoderUser.UserID)
	}
}

func TestLoadFromSystem_HomeEnvError(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping darwin-specific test on non-darwin")
	}
	origHome := os.Getenv("HOME")
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Unsetenv("HOME")
	os.Unsetenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("HOME", origHome)
		if origXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	}()

	_, err := LoadFromSystem()
	if err == nil {
		t.Fatal("expected error when HOME is unset, got nil")
	}
	t.Logf("got expected error: %v", err)
}

func TestLoadFromSystem_InvalidDatabase(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping darwin-specific test on non-darwin")
	}
	tmpDir := t.TempDir()

	dbDir := filepath.Join(tmpDir, "Library", "Application Support",
		"AgnesCode", "User", "globalStorage")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("failed to create db directory: %v", err)
	}

	dbPath := filepath.Join(dbDir, "state.vscdb")
	if err := os.WriteFile(dbPath, []byte("this is not a sqlite database"), 0644); err != nil {
		t.Fatalf("failed to write fake database: %v", err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	_, err := LoadFromSystem()
	if err == nil {
		t.Fatal("expected error for invalid database file, got nil")
	}
	t.Logf("got expected error: %v", err)
}

func TestLoadFromSystem_ValidDatabase(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping darwin-specific test on non-darwin")
	}
	tmpDir := t.TempDir()

	createTestDB(t, tmpDir, `{"agnesCoderUser":{"JWTToken":"valid-jwt-token-abc","userId":"valid-user-xyz"}}`)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	creds, err := LoadFromSystem()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if creds.JWTToken != "valid-jwt-token-abc" {
		t.Errorf("JWTToken = %q, want %q", creds.JWTToken, "valid-jwt-token-abc")
	}
	if creds.UserID != "valid-user-xyz" {
		t.Errorf("UserID = %q, want %q", creds.UserID, "valid-user-xyz")
	}
}

func TestLoadFromSystem_DatabaseMissingKey(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping darwin-specific test on non-darwin")
	}
	tmpDir := t.TempDir()

	dbDir := filepath.Join(tmpDir, "Library", "Application Support",
		"AgnesCode", "User", "globalStorage")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("failed to create db directory: %v", err)
	}

	dbPath := filepath.Join(dbDir, "state.vscdb")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to create sqlite database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE ItemTable (key TEXT PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("failed to create ItemTable: %v", err)
	}
	_, err = db.Exec("INSERT INTO ItemTable (key, value) VALUES ('some.other.key', 'irrelevant')")
	if err != nil {
		t.Fatalf("failed to insert row: %v", err)
	}
	db.Close()

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	_, err = LoadFromSystem()
	if err == nil {
		t.Fatal("expected error when AgnesCoder.IDE key is missing, got nil")
	}
	t.Logf("got expected error: %v", err)
}

func TestLoadFromSystem_DatabaseInvalidJSON(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping darwin-specific test on non-darwin")
	}
	tmpDir := t.TempDir()

	createTestDB(t, tmpDir, `{not valid json!!!}`)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	_, err := LoadFromSystem()
	if err == nil {
		t.Fatal("expected error for invalid JSON in database, got nil")
	}
	t.Logf("got expected error: %v", err)
}

func createTestDB(t *testing.T, baseDir string, jsonValue string) {
	t.Helper()

	dbDir := filepath.Join(baseDir, "Library", "Application Support",
		"AgnesCode", "User", "globalStorage")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("failed to create db directory: %v", err)
	}

	dbPath := filepath.Join(dbDir, "state.vscdb")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to create sqlite database: %v", err)
	}

	_, err = db.Exec("CREATE TABLE ItemTable (key TEXT PRIMARY KEY, value TEXT)")
	if err != nil {
		db.Close()
		t.Fatalf("failed to create ItemTable: %v", err)
	}

	_, err = db.Exec("INSERT INTO ItemTable (key, value) VALUES ('AgnesCoder.IDE', ?)", jsonValue)
	if err != nil {
		db.Close()
		t.Fatalf("failed to insert test data: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("failed to close database after setup: %v", err)
	}
}
