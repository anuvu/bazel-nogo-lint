package main

import (
	"crypto/md5"
	"fmt"
	"net/http"
)

func ReturnId() (id int, err error) {
	mySlice := make([]struct{}, 5)
	for i := 0; i != 5; i++ {
		mySlice = append(mySlice, struct{}{})
	}
	id = 10
	return
}

//gochecknoglobals
var g int = 21

var s = ""

func init() {
	g = 0

	s = "receved this is a very very very very very very very very very long long long long long line line line line line line but can get longer"

}

func main() {
	htt := "http"
	htt = "https"
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
