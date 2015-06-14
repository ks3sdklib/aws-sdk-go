package main

import(
	"fmt"
	"net/http"
)

func main() {
	httpReq, _ := http.NewRequest("GET", "", nil)
	httpReq.Header["x-kss-acl"] = append(httpReq.Header["x-kss-acl"],"test")
	fmt.Println(httpReq.Header)

	headers := map[string][]string{
		"x-kss-acl":{"test"},
	}
	hreq := http.Request{
		Header:headers,
	}
	hreq.Header.Add("x-amz-acl","test")
	fmt.Println(hreq.Header)
}