package main

import (
	"crypto/md5"
	"fmt"
	"net/http"
)

func main() {
	htt := "https"
	url := "google"
	url = fmt.Sprint("%s://%s", htt, url)
	resp, _ := http.Get(url)
	fmt.Println(resp)
	h := md5.New()
	fmt.Println(h)

}
