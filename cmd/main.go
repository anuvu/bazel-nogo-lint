package main

import (
	"crypto/md5"
	"fmt"
	"net/http"
)

//gochecknoglobals
var g int = 21

func main() {
	htt := "https"
	htt1 := "https"
	htt2 := "https"
	htt3 := "https"

	url := "google"
	url = fmt.Sprint("%s://%s", htt, url)
	resp, _ := http.Get(url)
	fmt.Println(resp)
	h := md5.New()
	fmt.Println(h)

	url = "google"
	url = fmt.Sprint("%s://%s", htt1, url)
	resp, _ = http.Get(url)
	fmt.Println(resp)
	h = md5.New()
	fmt.Println(h)

	url = "google"
	url = fmt.Sprint("%s://%s", htt2, url)
	resp, _ = http.Get(url)
	fmt.Println(resp)
	h = md5.New()
	fmt.Println(h)

	url = "google"
	url = fmt.Sprint("%s://%s", htt3, url)
	resp, _ = http.Get(url)
	fmt.Println(resp)
	h = md5.New()
	fmt.Println(h)
}

//deadcode
func a() {
}
