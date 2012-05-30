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
import "github.com/kless/goconfig/config"
import "encoding/json"
//import "os"
import "io/ioutil"

BUILD_NUMBER = "1.0._BUILD_"

type jsonobject struct {
    Object ObjectType
}

type ObjectType struct {
    Buffer_size int
    Databases   []DatabasesType
}

type DatabasesType struct {
}

func postData(data string, log_filename string) {
//    HOST := "localhost:4567"
    HOST := "ec2-107-22-44-241.compute-1.amazonaws.com:8086"

    API_KEY := "ignored"
    SERVER_NAME := "bobs_server"
    log_id := 123
//    contents,_ := ioutil.ReadAll(data);
//	buf := bytes.NewBuffer("your string")
	buf2 := bytes.NewBufferString(data)
    //TODO url escaping
    url := fmt.Sprintf("http://%s/api/v1/logs/%s/agents/%s?api_key=%s", HOST, log_id, SERVER_NAME, API_KEY)
    fmt.Printf("posting to url -%s\n%s\n", url,buf2)
    http.Post(url, "application/text", buf2)
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
            postData(sbuffer, filename)
            sbuffer = ""
            lines = 0
        }
        if err != nil {
            log.Fatal(err)
        }
    }
}

func main() {
    fmt.Printf("ERRPlane Local Agent starting, Version %s \n", BUILD_NUMBER)
    c, _ := config.ReadDefault("samples/sample_base_config")
    api_url,_ := c.String("DEFAULT", "host")
    fmt.Printf("----%s---\n", api_url)

/*
    file, e := ioutil.ReadFile("samples/sample_config.json")
    if e != nil {
        fmt.Printf("File error: %v\n", e)
        os.Exit(1)
    }
    fmt.Printf("%s\n", string(file))
*/  
    resp, err := http.Get(api_url + "/json_api")
    if err != nil {
        // handle error
        fmt.Printf("error getting config data-%s\n",err) 
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)

    var jsontype jsonobject
    json.Unmarshal(body, &jsontype)
    fmt.Printf("Results: %v\n", jsontype)

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
