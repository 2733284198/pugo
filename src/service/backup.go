package service

import (
	"errors"
	"fmt"
	"github.com/Unknwon/cae/zip"
	"github.com/Unknwon/com"
	"github.com/fuxiaohei/pugo/src/core"
	"github.com/fuxiaohei/pugo/src/model"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"
)

var (
	Backup = new(BackupService)

	ErrBackupDoing = errors.New("backup-is-doing")
)

type BackupService struct {
	IsDoingBackup bool
}

type BackupOption struct {
	WithBasic  bool
	WithData   bool
	WithStatic bool
	WithTheme  bool
}

func (bs *BackupService) Backup(v interface{}) (*Result, error) {
	if bs.IsDoingBackup {
		return nil, ErrBackupDoing
	}
	opt, ok := v.(BackupOption)
	if !ok {
		return nil, ErrServiceFuncNeedType(bs.Backup, opt)
	}
	fileName := fmt.Sprintf("%s/%s.zip", core.BackupDirectory, time.Now().Format("20060102150405"))
	dirName := path.Dir(fileName)
	if !com.IsDir(dirName) {
		if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
			return nil, err
		}
	}
	fileWriter, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_SYNC, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer fileWriter.Close()
	zip.Verbose = false
	zipWriter := zip.New(fileWriter)
	defer zipWriter.Close()
	if opt.WithBasic {
		zipWriter.AddFile("config.ini", "./config.ini")
	}
	if opt.WithData {
		zipWriter.AddDir("data", "./data")
	}
	if opt.WithStatic {
		zipWriter.AddDir("static", "./static")
	}
	if opt.WithTheme {
		zipWriter.AddDir("theme", "./theme")
	}
	if err := zipWriter.Flush(); err != nil {
		return nil, err
	}
	return newResult(bs.Backup, &fileName), nil
}

func (bs *BackupService) Files(_ interface{}) (*Result, error) {
	files := make([]*model.BackupFile, 0)
	if err := filepath.Walk(core.BackupDirectory, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".zip" {
			return nil
		}
		m := &model.BackupFile{
			Name:       filepath.Base(path),
			FullPath:   path,
			Size:       info.Size(),
			CreateTime: info.ModTime().Unix(),
		}
		files = append(files, m)
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Sort(model.BackupFiles(files))
	return newResult(bs.Files, &files), nil
}
