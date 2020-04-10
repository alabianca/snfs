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

func (n *Node) Update(nc snfsd.NodeConfiguration) error {
	return n.update(&nc)
}

func (n *Node) Delete(id int) error {
	return nil
}

func (n *Node) update(nc *snfsd.NodeConfiguration) error {
	stmt, err := n.DB.Prepare("UPDATE nodes SET cport=?, dport=?,name=?,processId=?,started=?, nodeId=? WHERE id = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(nc.Cport, nc.Dport, nc.Name, nc.ProcessId, nc.Started, nc.NodeId, nc.ID)
	return err
}

func (n *Node) create(nc *snfsd.NodeConfiguration) error {
	stmt, err := n.DB.Prepare("INSERT INTO nodes (cport, dport, name, started) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}

	result, err := stmt.Exec(nc.Cport, nc.Dport, nc.Name, 0)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	nc.ID = id
	return err
}

func createNodeTableIfNotExist(db *sql.DB) (*sql.Stmt, error) {
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS nodes (id INTEGER PRIMARY KEY, cport INTEGER NOT NULL, dport INTEGER NOT NULL, name TEXT DEFAULT '', processId INTEGER DEFAULT 0, started INTEGER DEFAULT 0, nodeId TEXT DEFAULT NULL)")
	if err != nil {
		return stmt, err
	}

	_, err = stmt.Exec()
	return stmt, err
}