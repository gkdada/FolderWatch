package scanner

import (
	"fmt"
	"folderwatch/config"
	"folderwatch/dbsqlite"
	"folderwatch/fwlog"
	"io"
	"os"

	"github.com/radovskyb/watcher"
)

type Scanner struct {
	cfg *config.Configs
	chn <-chan watcher.Event
	dbs *dbsqlite.DatabaseHolder
	lf  *fwlog.LogFile
}

func CreateScanner(cnf *config.Configs, channel <-chan watcher.Event, dh *dbsqlite.DatabaseHolder, lfl *fwlog.LogFile) *Scanner {
	return &Scanner{
		cfg: cnf,
		chn: channel,
		dbs: dh,
		lf:  lfl,
	}
}

func (scn *Scanner) HandleFileSystemChange() {
	for {
		fsc, ok := <-scn.chn
		if !ok {
			break
		}
		//handle operation.
		switch fsc.Op {
		case watcher.Create:
			scn.dbs.AddRecord(&fsc)
		case watcher.Write:
			scn.dbs.UpdateRecord(&fsc)
		case watcher.Chmod:
			scn.dbs.UpdateRecord(&fsc)
		case watcher.Remove:
			scn.dbs.DeleteRecord(&fsc)
		case watcher.Move:
		case watcher.Rename:
		}
	}
}

func (scn *Scanner) AddInitialRecords() error {
	fpr, err := os.Open(scn.cfg.TargetDir)
	if err != nil {
		scn.lf.LogError("error opening target directory", err)
		return err
	}
	defer fpr.Close()
	flst, err := fpr.ReadDir(0)
	if err != nil && err != io.EOF {
		scn.lf.LogError("error reading target directory", err)
		return err
	}
	fmt.Println("adding", len(flst), "entries to database")
	for _, di := range flst { //for now, we're skipping subfolders. TODO: traverse the subfolder recursively and add those files as well.
		if di.Type().IsRegular() {
			fi, err := di.Info()
			if err == nil {
				fullPath := fmt.Sprintf("%s%c%s", scn.cfg.TargetDir, os.PathSeparator, fi.Name())
				err := scn.dbs.AddInitialRecord(fullPath, di.Name(), fi.Mode().String(), fi.Size(), fi.ModTime())
				if err != nil {
					return err
				}
			} else {
				scn.lf.LogError("Error retrieving info for "+di.Name(), err)
			}
		}
	}
	return nil
}
