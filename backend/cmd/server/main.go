package main

import (
	"log"
	"os"

	"github.com/nefu-dev/wx-note/internal/api"
	"github.com/nefu-dev/wx-note/internal/repository"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8100"
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	os.MkdirAll(dataDir, 0755)

	db, err := repository.InitDB(dataDir)
	if err != nil {
		log.Fatalf("database init failed: %v", err)
	}
	defer db.Close()

	handler := api.NewHandler(db)
	r := handler.Setup()

	log.Printf("wx_note server starting on 0.0.0.0:%s", port)
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
