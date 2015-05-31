/*
 * sub-downloader.go
 * Copyright (C) 2015 Yang Junyong <yanunon@gmail.com>
 *
 * Distributed under terms of the MIT license.
 */

package main

import (
	"flag"
	"github.com/yanunon/shooter"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	is_print_log bool = false
)

func print_log(format string, v ...interface{}) {
	if is_print_log {
		log.Printf(format, v...)
	}
}

func downloadSubForDir(dir string) {
	print_log("start download sub for dir:%s\n", dir)
	fi, err := os.Stat(dir)
	if err != nil || !fi.IsDir() {
		print_log("%s is not an existed dir\n", dir)
	}
	filepath.Walk(dir, downloadSubWalkFunc)
}

func downloadSubWalkFunc(fpath string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		return nil
	}

	exts := []string{"avi", "mp4", "mkv", "mov", "mpeg", "wmv", "flv"}
	extIdx := strings.LastIndex(fpath, ".")
	if extIdx > -1 {
		ext := fpath[extIdx+1:]
		ext = strings.ToLower(ext)
		for _, e := range exts {
			if ext == e {
				downloadSubForFile(fpath)
				break
			}
		}
	}
	return nil
}

func downloadSubForFile(fpath string) {
	print_log("start download for file :%s\n", fpath)
	subtitles, err := shooter.GetSubtitlesInfo(fpath, "chn")
	if err != nil {
		print_log("download for file :%s error :%s\n", fpath, err.Error())
		return
	}
	for _, subtitle := range subtitles {
		err = shooter.DownloadSubtitle(subtitle, fpath)
		if err != nil {
			print_log("download for file :%s error :%s\n", fpath, err.Error())
		}
	}
}

func main() {
	var print_log = flag.Bool("log", false, "display log")
	var ignore_ssl_error = flag.Bool("ise", false, "ignore ssl error")
	var dir_path = flag.String("r", "", "download subtitle for dir")
	var file_path = flag.String("f", "", "download subtitle for file")
	flag.Parse()

	if *ignore_ssl_error {
		shooter.SetIgnoreSSLError(true)
	}

	is_print_log = *print_log

	if *dir_path != "" {
		downloadSubForDir(*dir_path)
	} else if *file_path != "" {
		downloadSubForFile(*file_path)
	} else {
		flag.Usage()
	}
}
