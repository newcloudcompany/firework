package ipam

import (
	"database/sql"
	"fmt"
	"net/netip"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type IPAM struct {
	db *sql.DB
}

func NewIPAM(dbPath string, cidr string) (*IPAM, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// Create IPAM database if it doesn't exist
		db, err := createIpamSqliteDatabase(dbPath, cidr)
		if err != nil {
			return nil, err
		}

		return &IPAM{
			db: db,
		}, nil
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database %s: %w", dbPath, err)
	}

	return &IPAM{
		db: db,
	}, nil
}

func createIpamSqliteDatabase(pathname string, cidr string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", pathname)
	if err != nil {
		return nil, err
	}

	// Create a table
	schema := `
	CREATE TABLE ips (
		addr TEXT PRIMARY KEY,
		is_free INTEGER,
		hostname TEXT
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	// Define the CIDR block
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return nil, err
	}

	// Insert all IP addresses in the CIDR block into the database
	for ip := prefix.Addr().Next().Next(); prefix.Contains(ip); ip = ip.Next() {
		// Transform the IP into the CIDR so that the string representation has the slash suffix
		addr := fmt.Sprintf("%s/%d", ip.String(), prefix.Bits())
		// log.Printf("Executing: INSERT INTO ips (addr, is_free) VALUES (%s, %d)\n", addr, 1)
		_, err := db.Exec("INSERT INTO ips (addr, is_free) VALUES (?, ?)", addr, 1)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func (ipam *IPAM) AllocateFreeIPAddress(hostname string) (string, error) {
	// Find a free IP address
	var addr string
	err := ipam.db.QueryRow("SELECT addr FROM ips WHERE is_free = 1 LIMIT 1").Scan(&addr)
	if err != nil {
		return "", err
	}

	// Mark the IP address as used
	_, err = ipam.db.Exec("UPDATE ips SET is_free = 0, hostname = ? WHERE addr = ?", hostname, addr)
	if err != nil {
		return "", err
	}

	return addr, nil
}
