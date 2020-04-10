package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/alabianca/snfs/snfsd"
	"github.com/alabianca/snfs/snfsd/http"
	"github.com/alabianca/snfs/snfsd/http/chi"
	"github.com/alabianca/snfs/snfsd/node"
	"github.com/alabianca/snfs/snfsd/pubsub"
	"github.com/alabianca/snfs/snfsd/server"
	"github.com/alabianca/snfs/snfsd/sqlite"
	"github.com/alabianca/snfs/snfsd/watchdog"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mitchellh/go-homedir"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"
)


func main() {
	db := getDB()
	defer db.Close()
	ps := pubsub.NewPubSub()

	pr, pw := io.Pipe()

	wd := watchdog.New(ps, pw)

	go func() {
		for {
			reader := bufio.NewReader(pr)
			line, _, _ := reader.ReadLine()
			fmt.Println("---------------")
			fmt.Println(string(line))
		}
	}()

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
		Port:     getPort(8080),
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

func getDB() (*sql.DB) {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	dbDir := path.Join(home, ".snfs", "snfs.db")

	db, err := sql.Open("sqlite3", dbDir)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func getPort(def int) int {
	port, err := strconv.ParseInt(os.Getenv("SNFSD_PORT"), 10, 16)
	if err != nil || port == 0 {
		return def
	}

	return int(port)
}
