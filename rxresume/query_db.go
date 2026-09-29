package main

import (
	"context"
	"fmt"
	"os"
	"github.com/jackc/pgx/v5"
)

func main() {
	conn, err := pgx.Connect(context.Background(), "postgresql://neondb_owner:npg_LGc5eD9PTNpV@ep-nameless-base-ayly8qxk-pooler.c-5.us-east-2.aws.neon.tech/rx-resume-db?sslmode=require")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	var data string
	err = conn.QueryRow(context.Background(), "SELECT data FROM resumes WHERE user_id = $1", "usr-anGBnlpTJ1").Scan(&data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(data)
}
