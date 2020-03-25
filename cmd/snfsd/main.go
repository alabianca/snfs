package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/alabianca/snfs/snfsd"
	"github.com/alabianca/snfs/snfsd/http/chi"
	"github.com/alabianca/snfs/snfsd/node"
	"github.com/alabianca/snfs/snfsd/pubsub"
	"github.com/alabianca/snfs/snfsd/server"
	"github.com/alabianca/snfs/snfsd/sqlite"
	"github.com/alabianca/snfs/snfsd/watchdog"
	"github.com/alabianca/snfs/snfsd/http"
	"log"
	"os"
	"os/signal"
	"syscall"
)


func main() {
	db := getDB()
	ps := pubsub.NewPubSub()
	wd := watchdog.New(ps)

	nodeDal := sqlite.Node{DB: db}
	nodeService := node.NodeService{
		Node:      &nodeDal,
		Publisher: ps,
	}

	appCtx := snfsd.AppContext{
		PubSub:      ps,
		NodeService: &nodeService,
	}


	srv := server.Server{
		Watchdog: wd,
		Handler:  http.App(&appCtx, chi.Routes),
		Host:     "",
		Port:     8080,
	}

	exit := make(chan struct{})
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-done
		log.Println("Shutting down server ...")
		close(exit)
	}()

	if err := srv.Run(exit); err != nil {
		log.Println(err)
	}
}

func getDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		log.Fatal(err)
	}

	return db
}
