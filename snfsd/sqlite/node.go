package sqlite

import (
	"database/sql"
	"github.com/alabianca/snfs/snfsd"
)

type Node struct {
	DB *sql.DB
}

func (n *Node) Create(nc *snfsd.NodeConfiguration) error {
	_, err := createNodeTableIfNotExist(n.DB)
	if err != nil {
		return err
	}

	return n.create(nc)
}

func (n *Node) Delete(id int) error {
	return nil
}

func (n *Node) create(nc *snfsd.NodeConfiguration) error {
	stmt, err := n.DB.Prepare("INSERT INTO nodes (cport, dport, fport, name) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}

	_ , err = stmt.Exec(nc.Cport, nc.Dport, nc.Fport, nc.Name)
	return err
}

func createNodeTableIfNotExist(db *sql.DB) (*sql.Stmt, error) {
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS nodes (id INTEGER PRIMARY KEY, cport INTEGER, dport INTEGER, fport INTEGER, name TEXT, processId INTEGER )")
	if err != nil {
		return stmt, err
	}

	_, err = stmt.Exec()
	return stmt, err
}