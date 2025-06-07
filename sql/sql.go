package sql

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p/core/crypto"
	_ "github.com/mattn/go-sqlite3"
)

type GetPodsStruct struct {
	PodName       string
	InternalPort  int
	Metadata      []string
	Images        []string
	ExternalImage string
}

type UserStruct struct {
	CID       string
	RoleName  string
	CreatedAt string
}

// var db *sql.DB

// type App struct {
// 	DB *sql.DB
// }

func SQLgetDB() (*sql.DB, error) {

	var err error
	db, err := sql.Open("sqlite3", "./conductor.db")
	if err != nil {
		return nil, err
	}

	// Connect check
	err = db.Ping()

	return db, err

}

// Open a connection to the conductor database.
// If there is no database, create a default database
func SQLinitDB() (*sql.DB, error) {

	var err error
	db, err := sql.Open("sqlite3", "./conductor.db")
	if err != nil {
		return nil, err
	}

	// Connect check
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	dht := generateRandomString(15)
	privkey, _ := generateKey()
	// Creating a table (if it does not exist)
	createTableSQL := `CREATE TABLE IF NOT EXISTS pods (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		PodName TEXT,
		InternalPort INTEGER,
		Images TEXTJ,
		ExternalImage TEXT,
		Hash TEXT UNIQUE,
		Metadata TEXTJ
	);
	
	CREATE TABLE IF NOT EXISTS roles (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    RoleName TEXT UNIQUE NOT NULL
	);

	CREATE TABLE IF NOT EXISTS users (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    Role INTEGER,
    CID TEXT,
    CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (Role) REFERENCES roles(Id)
	);

	CREATE TABLE IF NOT EXISTS settings (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    Port INTEGER,
    DHT TEXT,
	PrivKey BLOB,
	Version INTEGER,
    CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	INSERT OR IGNORE INTO roles (RoleName) VALUES
	('admin'),
	('user'),
	('guest');

	INSERT OR IGNORE INTO settings (Id, Port, DHT, PrivKey, Version) VALUES
	(1, 41537, ?, ?, 1)`
	_, err = db.Exec(createTableSQL, dht, privkey)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Function for adding a Pod
func SQLaddPod(db *sql.DB, PodName string, InternalPort int, Images []string, Metadata []string, Hash string, ExternalImage string) error {

	jsonData, err := json.Marshal(Metadata)
	if err != nil {
		return fmt.Errorf("SQLaddPod> %w", err)

	}

	jsonDataImg, err := json.Marshal(Images)
	if err != nil {
		return fmt.Errorf("SQLaddPod> %w", err)
	}

	insertSQL := `INSERT INTO pods (PodName, InternalPort, Images, Hash, Metadata, ExternalImage) VALUES (?, ?, ?, ?, ?, ?)`
	_, err = db.Exec(insertSQL, PodName, InternalPort, jsonDataImg, Hash, jsonData, ExternalImage)
	return err
}

func SQLgetPods(db *sql.DB, hash string) (GetPodsStruct, error) {

	if db == nil {
		return GetPodsStruct{}, fmt.Errorf("SQLgetPods No connection to the database")
	}

	type Pod struct {
		Images        json.RawMessage `json:"images"`
		Metadata      json.RawMessage `json:"metadata"`
		InternalPort  int
		PodName       string
		ExternalImage string
	}

	var pod Pod
	err := db.QueryRow("SELECT Images, Metadata, InternalPort, PodName, ExternalImage FROM pods WHERE Hash = $1", hash).Scan(&pod.Images, &pod.Metadata, &pod.InternalPort, &pod.PodName, &pod.ExternalImage)
	if err != nil {
		return GetPodsStruct{}, err
	}

	var images []string
	var metadata []string

	err = json.Unmarshal(pod.Images, &images)
	if err != nil {
		return GetPodsStruct{}, err
	}
	err = json.Unmarshal(pod.Metadata, &metadata)
	if err != nil {
		return GetPodsStruct{}, err
	}

	// Check if there is data in the structure
	if pod.PodName == "" || len(images) == 0 {
		return GetPodsStruct{}, fmt.Errorf("no data found for hash: %s", hash)
	}

	return GetPodsStruct{PodName: pod.PodName, InternalPort: pod.InternalPort, Metadata: metadata, Images: images, ExternalImage: pod.ExternalImage}, nil

}

func SQLaddUser(db *sql.DB, role int, sid string) error {
	insertSQL := `INSERT INTO users (Role, CID) VALUES (?, ?)`
	_, err := db.Exec(insertSQL, role, sid)
	return err
}

func SQLdeleteUser(db *sql.DB, role int, sid string) error {
	query := "DELETE FROM users WHERE Role = ? AND CID = ?"

	_, err := db.Exec(query, role, sid)
	if err != nil {
		return err
	}
	return nil
}

func SQLlistUsers(db *sql.DB) ([]UserStruct, error) {

	var users []UserStruct
	rows, err := db.Query(`
		SELECT u.CID, r.RoleName, u.CreatedAt
		FROM users u
		JOIN roles r ON u.Role = r.Id;
	`)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user UserStruct
		if err := rows.Scan(&user.CID, &user.RoleName, &user.CreatedAt); err != nil {
			return users, err
		}
		users = append(users, user)
	}

	return users, nil

}

func SQLcheckRole(db *sql.DB, sid string) (int, error) {
	var role int
	err := db.QueryRow("SELECT Role FROM Users WHERE CID = $1", sid).Scan(&role)
	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return role, nil
}

func SQLGetAllPods(db *sql.DB) ([]byte, error) {

	rows, err := db.Query("SELECT PodName, Hash FROM Pods")
	if err != nil {
		return nil, fmt.Errorf("SQLGetAllPods>db.Query error: %w", err)
	}

	type Pod struct {
		PodName string `xml:"PodName"`
		Hash    string `xml:"Hash"`
	}

	type Response struct {
		XMLName xml.Name `xml:"Response"`
		Pods    []Pod    `xml:"Pod"`
	}

	var pods []Pod

	for rows.Next() {
		var pod Pod
		if err := rows.Scan(&pod.PodName, &pod.Hash); err != nil {
			return nil, fmt.Errorf("SQLGetAllPods>rows.Scan error: %w", err)
		}
		pods = append(pods, pod)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("SQLGetAllPods>rows.Err error: %w", err)
	}

	response := Response{Pods: pods}

	xmlData, err := xml.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("SQLGetAllPods>xml.MarshalIndent error: %w", err)
	}
	return xmlData, nil
}

func SQLgetSettings(db *sql.DB) (int, string, crypto.PrivKey, error) {
	var port int
	var dht string
	var PrivKey []byte
	err := db.QueryRow("SELECT Port, DHT, PrivKey  FROM settings WHERE id = 1").Scan(&port, &dht, &PrivKey)
	if err != nil {
		return port, dht, nil, err
	}

	privKeyRSA, err := crypto.UnmarshalPrivateKey(PrivKey)
	if err != nil {
		return port, dht, nil, err
	}

	return port, dht, privKeyRSA, nil
}

func SQLdeletePod(db *sql.DB, hash string) error {
	query := "DELETE FROM pods WHERE Hash = ?"

	_, err := db.Exec(query, hash)
	if err != nil {
		return err
	}
	return nil
}
