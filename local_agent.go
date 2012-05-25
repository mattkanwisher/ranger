package main

import "fmt"
import "net/http"
import "bytes"
//import "io/ioutil"
import "log"
import "os/exec"
//import "io"
import "bufio"
import "time"
import "runtime"

func postData(data string) {
//    contents,_ := ioutil.ReadAll(data);
//	buf := bytes.NewBuffer("your string")
	buf2 := bytes.NewBufferString(data)
    http.Post("http://localhost:4567/hi", "application/text", buf2)
//TODO HANDLE ERROR AND RETRIES!
}


func readData(filename string) {
    fmt.Printf("in read data !")
    cmd := exec.Command("tail", "-f", filename)
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        log.Fatal(err)
    }
    if err := cmd.Start(); err != nil {
        log.Fatal(err)
    }

    err  = nil
    reader := bufio.NewReader(stdout)
    sbuffer := ""
    lines := 0
    for ; err == nil;  {
        s,err := reader.ReadString('\n')
        fmt.Printf("got line-%s---%s-\n", filename, s)
        sbuffer += s
        lines += 1
        if(lines > 5 ) { //|| time > 1 min) {
            fmt.Printf("Clearing buffer and posting to http\n")
            postData(sbuffer)
            sbuffer = ""
            lines = 0
        }
        if err != nil {
            log.Fatal(err)
        }
    }
}

func main() {
    fmt.Printf("hello, world\n")

 
    path, err := exec.LookPath("tail")
    if err != nil {
        log.Fatal("installing tail is in your future")
//        exit(1)
    }
    fmt.Printf("tail is available at %s\n", path)

    filename := "test_logs/test_log.log"
    go readData(filename)
    filename2 := "test_logs/test_log2.log"
    go readData(filename2)

    if err != nil {
        log.Fatal(err)
    }

    err = nil
    for ; err == nil;  {
        //TODO monitor go routines, if one exists reload it
        time.Sleep(0)
        runtime.Gosched()
    }
//    postData(stdout)
}
