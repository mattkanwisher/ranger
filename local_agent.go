package main

import "fmt"
import "net/http"
import "bytes"
import "io/ioutil"

func main() {
    fmt.Printf("hello, world\n")
	contents,_ := ioutil.ReadFile("test_logs/test_log.log");
//	buf := bytes.NewBuffer("your string")
	buf2 := bytes.NewBuffer(contents)
    //resp, err :=
    http.Post("http://localhost:4567/hi", "application/text", buf2)
}
