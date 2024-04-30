package dbsqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"folderwatch/fwlog"
	"folderwatch/types"
	"log"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DatabaseHolder struct {
	dbs *sql.DB
	lf  *fwlog.LogFile
	mtx sync.Mutex //this mutex is used to avoid concurrency issues. Only one operation is allowed at a time for now
	//not sure whether sql/sqlite allows multiple operations concurrently.
}

func OpenDatabase(dbName string, lfl *fwlog.LogFile) *DatabaseHolder {
	var dh DatabaseHolder

	var err error
	dh.lf = lfl
	dh.dbs, err = sql.Open("sqlite3", dbName)
	if err != nil {
		log.Fatalln("error opening database", err.Error())
	}

	sql := `CREATE TABLE if not exists fsevents (
		file_path TEXT PRIMARY KEY,  
		file_name TEXT NOT NULL,
		mode      TEXT NOT NULL,
		size      INT
	 );`

	_, err = dh.dbs.Exec(sql)

	if err == nil {
		dh.lf.LogInfo("database table 'fsevents' exists or has been created", nil)
	}

	res, err := dh.dbs.Exec("DELETE FROM fsevents")
	if err == nil {
		numRows, _ := res.RowsAffected()
		dh.lf.LogInfo(fmt.Sprintf("%d rows deleted from 'fsevents' table", numRows), nil)
	} else {
		dh.lf.LogError("error deleting all entries from fsevents", err)
	}

	return &dh
}

func (dh *DatabaseHolder) AddInitialRecord(fullPath string, fName string, md string, sz int64, dt time.Time) error {
	dh.mtx.Lock()
	defer dh.mtx.Unlock()

	dh.lf.LogEvent(types.Initial, fullPath, fName, md, sz, dt)
	stmt, err := dh.dbs.Prepare("INSERT INTO fsevents(file_path, file_name, mode, size) values(?, ?, ?, ?)")
	if err != nil {
		//TODO: log the error
		dh.lf.LogError("error preparing the add initial record statement:", err)
		return errors.New("error preparing the add initial record statement")
	}
	_, err = stmt.Exec(fullPath, fName, md, sz)
	return err
}

func (dh *DatabaseHolder) AddRecord(fullPath string, fName string, md string, sz int64, dt time.Time) error {
	dh.mtx.Lock()
	defer dh.mtx.Unlock()

	dh.lf.LogEvent(types.Add, fullPath, fName, md, sz, dt)
	stmt, err := dh.dbs.Prepare("INSERT INTO fsevents(file_path, file_name, mode, size) values(?, ?, ?, ?)")
	if err != nil {
		//TODO: log the error
		dh.lf.LogError("error preparing the add record statement:", err)
		return errors.New("error preparing the add record statement")
	}
	_, err = stmt.Exec(fullPath, fName, md, sz)
	if err != nil {
		dh.lf.LogError("error adding the record for "+fullPath, err)
	}
	return err
}

func (dh *DatabaseHolder) UpdateRecord(fullPath string, fName string, md string, sz int64, dt time.Time) error {
	dh.mtx.Lock()
	defer dh.mtx.Unlock()

	dh.lf.LogEvent(types.Update, fullPath, fName, md, sz, dt)
	stmt, err := dh.dbs.Prepare("UPDATE fsevents SET mode = ?, size = ? WHERE file_path = ?")
	if err != nil {
		//TODO: log the error
		dh.lf.LogError("error preparing the add record statement", err)
		return errors.New("error preparing the update record statement")
	}
	_, err = stmt.Exec(md, sz, fullPath)
	if err != nil {
		dh.lf.LogInfo("error updating the record for "+fullPath+" trying add", err)
		return dh.AddRecord(fullPath, fName, md, sz, dt)
	}
	return nil
}
func (dh *DatabaseHolder) DeleteRecord(fullPath string, fName string, md string, sz int64, dt time.Time) error {
	dh.mtx.Lock()
	defer dh.mtx.Unlock()

	dh.lf.LogEvent(types.Delete, fullPath, fName, md, sz, dt)
	_, err := dh.dbs.Exec("DELETE from fsevents WHERE file_path = ?", fullPath)
	if err != nil {
		dh.lf.LogInfo("error deleting the record for "+fullPath, err)
		return dh.AddRecord(fullPath, fName, md, sz, dt)
	}
	return nil
}
