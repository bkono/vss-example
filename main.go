package main

import (
	"database/sql"
	"log"

	_ "github.com/asg017/sqlite-vss/bindings/go"
	_ "github.com/mattn/go-sqlite3"
)

// #cgo linux,amd64 LDFLAGS: -L./extensions -Wl,-undefined,dynamic_lookup -lstdc++
// #cgo darwin,amd64 LDFLAGS: -L./extensions -Wl,-undefined,dynamic_lookup -lomp
// #cgo darwin,arm64 LDFLAGS: -L/opt/homebrew/opt/libomp/lib -L./extensions -Wl,-undefined,dynamic_lookup -lomp
import "C"

func main() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var version, vector string
	err = db.QueryRow("SELECT vss_version(), vector_to_json(?)", []byte{0x00, 0x00, 0x28, 0x42}).Scan(&version, &vector)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("version=%s vector=%s\n", version, vector)
}
