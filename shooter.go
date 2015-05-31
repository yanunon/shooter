/*
 * shooter.go
 * Copyright (C) 2015 Yang Junyong <yanunon@gmail.com>
 *
 * Distributed under terms of the MIT license.
 */

// Download subtitle from shooter.cn
package shooter

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	BASE_URL string = "http://shooter.cn/api/subapi.php"
)

var (
	ignore_ssl_error bool         = false
	http_client      *http.Client = &http.Client{}
)

type SubtitleFile struct {
	Ext  string
	Link string
}

type SubtitleInfo struct {
	Desc  string
	Delay int
	Files []SubtitleFile
}

/*
 * SVplayer hash method https://docs.google.com/document/d/1w5MCBO61rKQ6hI5m9laJLWse__yTYdRugpVyz4RzrmM/preview
 */
func get_block_md5(file *os.File, offset, size int64) (md5str string, err error) {
	data := make([]byte, size)
	n, err := file.ReadAt(data, offset)
	if err != nil || int64(n) != size {
		return
	}
	md5str = fmt.Sprintf("%x", md5.Sum(data))
	return
}

func get_file_md5(fpath string) (md5str string, err error) {
	file, err := os.Open(fpath)
	if err != nil {
		return
	}
	defer file.Close()

	fileinfo, err := file.Stat()
	if err != nil || fileinfo.Size() < 8192 {
		return
	}

	offsets := [4]int64{4096, fileinfo.Size() / 3 * 2, fileinfo.Size() / 3, fileinfo.Size() - 8192}
	md5strs := make([]string, 4)
	for idx, offset := range offsets {
		md5str, _ := get_block_md5(file, offset, 4096)
		md5strs[idx] = md5str
	}

	md5str = strings.Join(md5strs, ";")
	return
}

func get_short_name(fpath string) (shortName string) {
	name := filepath.Base(fpath)
	name = strings.ToLower(name)
	split_words := []string{"blueray", "bluray", "dvdrip", "xvid", "cd1", "cd2", "cd3", "cd4", "cd5", "cd6", "vc1", "vc-1", "hdtv", "1080p", "720p", "1080i", "x264", "stv", "limited", "ac3", "xxx", "hddvd"}
	for _, word := range split_words {
		idx := strings.Index(name, word)
		if idx > -1 {
			name = name[:idx]
		}
	}

	empty_chars := "[].-#_=+<>,"
	for _, c := range empty_chars {
		name = strings.Replace(name, string(c), " ", -1)
	}
	shortName = name
	return
}

// Get subtitles info for the movie by fpath, lang can be "chn" or "eng"
func GetSubtitlesInfo(fpath, lang string) (subtitles []SubtitleInfo, err error) {
	fpath, err = filepath.Abs(fpath)
	if err != nil {
		return
	}

	filehash, err := get_file_md5(fpath)
	if err != nil || filehash == "" {
		return
	}
	shortname := get_short_name(fpath)
	form := url.Values{}
	form.Set("filehash", filehash)
	form.Set("shortname", shortname)
	form.Set("format", "json")
	if lang != "chn" {
		form.Set("lang", lang)
	}
	resp, err := http_client.PostForm(BASE_URL, form)

	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || len(body) < 64 {
		return
	}
	err = json.Unmarshal(body, &subtitles)
	return
}

// Download subtitles to the movie's dir with same name, fpath is the movie's path
func DownloadSubtitle(subtitle SubtitleInfo, fpath string) (err error) {
	fpath, _ = filepath.Abs(fpath)
	dir := filepath.Dir(fpath)
	name := filepath.Base(fpath)
	extIdx := strings.LastIndex(name, ".")
	if extIdx > -1 {
		name = name[:extIdx]
	}

	if subtitle.Files != nil {
		for _, file := range subtitle.Files {
			subpath := filepath.Join(dir, name+"."+file.Ext)
			if _, err := os.Stat(subpath); os.IsNotExist(err) {
				err = downloadFile(file.Link, subpath)
				if err != nil {
					return err
				}
			}
		}
	}

	return
}

func downloadFile(url, path string) (err error) {
	resp, err := http_client.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(path, body, 0622)
	return
}

// Set for ignore ssl error
func SetIgnoreSSLError(ignore bool) {
	if ignore == ignore_ssl_error {
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if ignore {
		http_client = &http.Client{
			Transport: tr,
		}
	} else {
		http_client = &http.Client{}
	}
}
