package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

var (
	ctx context.Context
	db  *sql.DB
)

func queryUser(id int) error {
	var username string
	var created time.Time
	err := db.QueryRowContext(ctx, "SELECT username, created_at FROM users WHERE id=?", id).Scan(&username, &created)
	return fmt.Errorf("accessing DB: %w", err)
}


func main() {
	log.Println("start")
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/test");
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	id := 10000
	err = queryUser(id)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("not found id: %d", id)
	}

	if err != nil {
		log.Fatal(err)
	}
	log.Println("end")
}