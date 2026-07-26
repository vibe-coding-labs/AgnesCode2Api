package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

// Credentials holds AgnesCode authentication data.
type Credentials struct {
	PtKey         string
	UserID        string
	ColorBaseURL  string
	MasterBaseURL string
	Tenant        string
	LoginType     string
	OrgFullName   string
}

type stateData struct {
	AgnesCoderUser struct {
		PtKey         string `json:"ptKey"`
		UserID        string `json:"userId"`
		ColorBaseURL  string `json:"colorBaseUrl"`
		MasterBaseURL string `json:"masterBaseUrl"`
		Tenant        string `json:"tenant"`
		LoginType     string `json:"loginType"`
		OrgFullName   string `json:"orgFullName"`
	} `json:"joyCoderUser"`
}

const (
	stateDBEnv       = "AGNESCODE_STATE_DB"
	containerStateDB = "/root/.agnescode-ide/state.vscdb"
)

// LoadFromSystem reads ptKey from local AgnesCode state database (macOS).
func LoadFromSystem() (*Credentials, error) {
	if dbPath := os.Getenv(stateDBEnv); dbPath != "" {
		return loadFromStateDB(dbPath)
	}
	if _, err := os.Stat(containerStateDB); err == nil {
		return loadFromStateDB(containerStateDB)
	}
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("auto credential extraction requires macOS AgnesCode IDE state; in Docker, mount state.vscdb to %s or set %s", containerStateDB, stateDBEnv)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	dbPath := filepath.Join(home,
		"Library", "Application Support",
		"AgnesCode", "User", "globalStorage", "state.vscdb")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("AgnesCode state database not found at %s\n  Please install and log in to AgnesCode IDE first", dbPath)
	}

	return loadFromStateDB(dbPath)
}

func loadFromStateDB(dbPath string) (*Credentials, error) {
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("AgnesCode state database not found at %s: %w", dbPath, err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("cannot open AgnesCode database: %w", err)
	}
	defer db.Close()

	var value string
	if err := db.QueryRow(
		"SELECT value FROM ItemTable WHERE key='AgnesCoder.IDE'",
	).Scan(&value); err != nil {
		return nil, fmt.Errorf("login info not found in database\n  Please log in to AgnesCode IDE first")
	}

	var data stateData
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return nil, fmt.Errorf("cannot parse login data from database: %w", err)
	}
	if data.AgnesCoderUser.PtKey == "" {
		return nil, fmt.Errorf("ptKey is empty in stored credentials\n  Please re-login to AgnesCode IDE")
	}
	if data.AgnesCoderUser.UserID == "" {
		return nil, fmt.Errorf("userId is empty in stored credentials\n  Please re-login to AgnesCode IDE")
	}
	return &Credentials{
		PtKey:         data.AgnesCoderUser.PtKey,
		UserID:        data.AgnesCoderUser.UserID,
		ColorBaseURL:  data.AgnesCoderUser.ColorBaseURL,
		MasterBaseURL: data.AgnesCoderUser.MasterBaseURL,
		Tenant:        data.AgnesCoderUser.Tenant,
		LoginType:     data.AgnesCoderUser.LoginType,
		OrgFullName:   data.AgnesCoderUser.OrgFullName,
	}, nil
}
