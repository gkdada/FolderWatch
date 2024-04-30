package fwlog

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"folderwatch/config"
	"folderwatch/types"

	"github.com/radovskyb/watcher"
)

type LogFile struct {
	file       *os.File
	fPath      string
	cnfg       *config.Configs
	nextFileAt time.Time
	mtx        sync.Mutex
}

func CreateLogFile(cnf *config.Configs) *LogFile {
	lft := &LogFile{
		cnfg: cnf,
	}
	lft.openNew()

	return lft
}

func (lf *LogFile) openNew() bool {
	if lf.file != nil && time.Now().Before(lf.nextFileAt) {
		return true
	}
	if lf.file != nil {
		lf.file.Close()
		lf.file = nil
	}

	tm := time.Now()
	lf.fPath = fmt.Sprintf("%s%c%04d%02d%02d-%02d.log", lf.cnfg.LogFilePath, os.PathSeparator,
		tm.Year(), tm.Month(), tm.Day(), tm.Hour())
	fl, err := os.OpenFile(lf.fPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("error opening log file", err.Error())
		lf.file = nil
		return false
	}
	lf.nextFileAt = time.Now().Add(time.Hour * time.Duration(lf.cnfg.NewLogFileHours))
	lf.file = fl

	return true
}

func (lf *LogFile) LogInitial(fullPath string, fName string, md string, sz int64, dt time.Time) {

	fmt.Println("Adding initial record for", fullPath)

	lf.mtx.Lock()
	defer lf.mtx.Unlock()

	lf.openNew()

	if lf.file == nil {
		return
	}
	ch := types.FileChangeLog{
		ChangeType:  "Initial",
		FilePath:    fullPath,
		Mode:        md,
		Size:        sz,
		LastUpdated: dt,
	}
	bytes, err := json.MarshalIndent(ch, " ", "\t")
	if err != nil {
		strerr := fmt.Sprintln("error marshaling FileChangeLog object", err.Error())
		lf.file.Write([]byte(strerr))
	} else {
		lf.file.Write(bytes)
		lf.file.Write([]byte("\r\n"))
	}
}

func (lf *LogFile) LogEvent(fc types.FChange, ev *watcher.Event) {

	switch fc {
	case types.Initial:
		fmt.Println("Adding initial record for", ev.Path)
	case types.Add:
		fmt.Println("Recording an 'Add' event for", ev.Path)
	case types.Update:
		fmt.Println("Recording an 'Update' event for", ev.Path)
	case types.Delete:
		fmt.Println("Recording a 'Remove' event for", ev.Path)
	}

	lf.mtx.Lock()
	defer lf.mtx.Unlock()

	lf.openNew()

	if lf.file == nil {
		return
	}
	ch := types.FileChangeLog{
		ChangeType:  fc.String(),
		FilePath:    ev.Path,
		Mode:        ev.Mode().String(),
		Size:        ev.Size(),
		LastUpdated: ev.ModTime(),
	}
	bytes, err := json.MarshalIndent(ch, " ", "\t")
	if err != nil {
		strerr := fmt.Sprintln("error marshaling FileChangeLog object", err.Error())
		lf.file.Write([]byte(strerr))
	} else {
		lf.file.Write(bytes)
		lf.file.Write([]byte("\r\n"))
	}
}

func (lf *LogFile) LogInfo(info string, errs error) {
	lf.logProgress("info", info, errs)
}

func (lf *LogFile) LogError(info string, errs error) {
	lf.logProgress("error", info, errs)
	//also print out errors on to stdout
	fmt.Println(info, errs.Error())
}

func (lf *LogFile) logProgress(tp string, info string, errs error) {
	lf.mtx.Lock()
	defer lf.mtx.Unlock()

	lf.openNew()

	if lf.file == nil {
		return
	}
	pr := types.LogProgress{
		Type: tp,
		Info: info,
	}
	if errs != nil {
		pr.Error = errs.Error()
	}
	bytes, err := json.MarshalIndent(pr, " ", "\t")
	if err != nil {
		strerr := fmt.Sprintln("error marshaling status", err.Error())
		lf.file.Write([]byte(strerr))
	} else {
		lf.file.Write(bytes)
		lf.file.Write([]byte("\r\n"))
	}
}
