package main

import (
	"fmt"
	"golango.cn/go-music/core"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var (
	SongsName string
	Indexs    string
)

func main() {

	pwd, _ := os.Getwd()
	dir := pwd + "/download/"

	os.MkdirAll(dir, os.ModePerm)

	for {

	A:
		fmt.Println("请输入歌曲名称：")

		fmt.Scanln(&SongsName)
		SongsName = strings.TrimSpace(SongsName)
		if len(SongsName) == 0 {
			goto A
		}
		SongsName := url.QueryEscape(SongsName)

		songs, err := core.Search(SongsName)

		if err != nil {
			fmt.Println(err.Error())
			goto A
		}

		if len(songs) == 0 {
			fmt.Println("没有找到歌曲信息")
			goto A
		}

		fmt.Println()
		fmt.Printf("******************找到%d首歌曲******************\n", len(songs))
		fmt.Println()
		for _, s := range songs {
			fmt.Printf("%d、%s\n", s.Index, s.Name)
		}
		fmt.Println("-1、返回")
		fmt.Println()
		fmt.Println("*************************************************")
		fmt.Println("请输入下载编号：")
	B:

		fmt.Scanln(&Indexs)

		indexs := strings.FieldsFunc(Indexs, func(r rune) bool {
			return r == ':' || r == '.' || r == ',' || r == '、' || r == '\\'
		})

		for _, v := range indexs {

			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				fmt.Println(err.Error())
				goto B
			}
			if i <= 0 {
				goto A
			}

			song, err := core.Find(int(i), songs)
			if err != nil {
				fmt.Println(err.Error())
				goto B
			}

			fmt.Println("准备下载", song)
			filename, err := core.DownLoad(song, dir)

			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("歌曲下载成功", filename)
			}

		}

	}
}
