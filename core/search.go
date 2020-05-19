package core

import (
	"errors"
	"fmt"
	htmlquery "github.com/antchfx/xquery/html"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/hu17889/go_spider/core/common/request"
	"github.com/hu17889/go_spider/core/downloader"

	. "github.com/ahmetb/go-linq"
)

type Songs struct {
	Name  string
	Md5   string
	Index int
}

//var r = regexp.MustCompile(`javascript:follow\('([^']*)'\)`)
var r = regexp.MustCompile(`\?v=(.*)`)

var headers = http.Header{
	"User-Agent":      []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36"},
	"Accept":          []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
	"Accept-Language": []string{"zh-CN,zh;q=0.9"},
	"Accept-Encoding": []string{"gzip, deflate"},
}

func Search(songsName string) ([]*Songs, error) {

	//User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36
	//Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9
	//Accept-Language: zh-CN,zh;q=0.9
	//Accept-Encoding: gzip, deflate

	url := fmt.Sprintf("http://mp34.butterfly.mopaasapp.com/?mp3=%s", songsName)

	//headers := make(http.Header)
	//headers.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36")
	//headers.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	//headers.Add("Accept-Language", "zh-CN,zh;q=0.9")
	//headers.Add("Accept-Encoding", "gzip, deflate")

	requ := request.NewRequest(url, "html", "", "GET", "", headers, nil, nil, nil)
	page := downloader.NewHttpDownloader().Download(requ)

	if !page.IsSucc() {
		return nil, errors.New(page.Errormsg())
	}

	root, _ := html.Parse(strings.NewReader(page.GetBodyStr()))
	nodes := htmlquery.Find(root, `//*[@id="wlsong"]/ul/li`)

	var songs []*Songs

	for i, n := range nodes {

		one := htmlquery.FindOne(n, `a`)
		name := strings.Replace(htmlquery.InnerText(one), " ", "", -1)
		onclick := htmlquery.SelectAttr(one, "href")

		submatch := r.FindAllStringSubmatch(onclick, -1)
		if len(submatch) > 0 && len(submatch[0]) > 1 {
			onclick = submatch[0][1]
		}

		songs = append(songs, &Songs{
			Name:  name,
			Md5:   onclick,
			Index: (i + 1),
		})

	}

	return songs, nil
}

func Find(index int, songs []*Songs) (*Songs, error) {

	if songs == nil || len(songs) == 0 {
		return nil, errors.New("没有找到歌曲信息")
	}

	if index > len(songs) {
		return nil, errors.New("序号错误，请重新输入")
	}

	one := From(songs).Where(func(i interface{}) bool {
		return i.(*Songs).Index == index
	}).First()

	if one == nil {
		return nil, errors.New("没有找到歌曲信息")
	}

	return one.(*Songs), nil

}

func DownLoad(song *Songs, dir string) (string, error) {

	url := fmt.Sprintf("http://mp34.butterfly.mopaasapp.com/?v=%s", song.Md5)

	cookie := &http.Cookie{
		Name:    "x09",
		Value:   "x09",
		Expires: time.Now().Add(time.Second * 3600),
	}

	requ := request.NewRequest(url, "html", "", "GET", "", headers, []*http.Cookie{cookie}, nil, nil)
	page := downloader.NewHttpDownloader().Download(requ)

	if !page.IsSucc() {
		return "", errors.New(page.Errormsg())
	}

	root, _ := html.Parse(strings.NewReader(page.GetBodyStr()))
	one := htmlquery.FindOne(root, `//*[@id="audio"]`)

	src := htmlquery.SelectAttr(one, "src")

	ext := path.Ext(src)

	resp, err := http.Get(src)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fileName := dir + song.Name + ext
	err = ioutil.WriteFile(fileName, bytes, os.ModePerm)
	if err != nil {
		return "", err
	}

	return fileName, nil
}
