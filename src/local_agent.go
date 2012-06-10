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
import "os"
import "io/ioutil"
import "github.com/droundy/goopt"
import "strings"

var BUILD_NUMBER = "1.0._BUILD_"

type AgentConfigType struct {
    Id int
    Version string
    Server string
    Configuration_interval int
    Name string
    Organization_id int
    Updated_at string
    Agent_logs   []AgentLogType
}

type AgentLogType struct {
    Agent_id int
    Created_at string
    Id int
    Log_id int
    Updated_at string
    Log LogType
}

type LogType struct {
    Id int
    Name string
    Path string
    Created_at string
}


func postData(api_key string, api_server string, data string, log_filename string) {
    SERVER_NAME := "bobs_server"
    log_id := 123
//    contents,_ := ioutil.ReadAll(data);
//	buf := bytes.NewBuffer("your string")
	buf2 := bytes.NewBufferString(data)
    //TODO url escaping
    url := fmt.Sprintf("http://%s/api/v1/logs/%s/agents/%s?api_key=%s", api_server, log_id, SERVER_NAME, api_key)
    fmt.Printf("posting to url -%s\n%s\n", url,buf2)
    http.Post(url, "application/text", buf2)
//TODO HANDLE ERROR AND RETRIES!
}


func getSysStats() {
    fmt.Printf("in read data !")
    myos :=  runtime.GOOS

    var cmd *exec.Cmd
    // OSX
    if myos == "darwin" {
     cmd = exec.Command("top", "-l", "1")
    } else {
      // LINUX
      cmd = exec.Command("top", "-n1", "-b")
    }

    stdout, err := cmd.StdoutPipe()
    if err != nil {
        log.Fatal(err)
    }
    if err := cmd.Start(); err != nil {
        log.Fatal(err)
    }

    err  = nil
    contents,_ := ioutil.ReadAll(stdout)

    log.Printf("top output -%s\n\n========\n%s\n", contents, myos)
}

func readData(api_key, api_server, filename string) {
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
            postData(api_key, api_server, sbuffer, filename)
            sbuffer = ""
            lines = 0
        }
        if err != nil {
            log.Fatal(err)
        }
    }
}
/*
func parseJsonFromFile() {
        file, e := ioutil.ReadFile("samples/sample_config.json")
    if e != nil {
        fmt.Printf("File error: %v\n", e)
        os.Exit(1)
    }
    fmt.Printf("%s\n", string(file))

}
*/
//TODO: IF this fails  read from disk, if that fails sleep until the server is back online
func parseJsonFromHttp(api_url string, api_key string) {
    server := "TEST"
    full_config_url := fmt.Sprintf(api_url + "/api/v1/agents/%s/config?api_key=%s", server, api_key)
    log.Printf("api url %s\n", full_config_url)
    resp, err := http.Get(full_config_url)
    if err != nil {
        // handle error
        fmt.Printf("error getting config data-%s\n",err)    
        os.Exit(1)
        return
    }

    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)

    body2 := []byte(strings.Replace(string(body), "null", "\"\"", -1))//Go doesn't handle nulls in json very well, lets just cheat

    fmt.Printf("json-%s\n", body2)
    var jsontype  AgentConfigType
    err = json.Unmarshal(body2, &jsontype)
    if err != nil {
        // handle error
        fmt.Printf("error parsing config data-%s\n",err)    
        os.Exit(1)
        return
    }

    fmt.Printf("Results: %v\n", jsontype)    
}

var config_file = goopt.String([]string{"-c", "--config"}, "", "config file")

func main() {
    fmt.Printf("ERRPlane Local Agent starting, Version %s \n", BUILD_NUMBER)

    goopt.Description = func() string {
        return "ERRPlane Local Agent."
    }
    goopt.Version = BUILD_NUMBER
    goopt.Summary = "ErrPlane Log and System Monitor"
    goopt.Parse(nil)

    var fconfig_file string
    fconfig_file = *config_file
    if(fconfig_file == "") {
        fconfig_file = "samples/sample_base_config"
    }

    log.Print("Loading config file ", *config_file, ".")

    c, _ := config.ReadDefault(fconfig_file)
    api_url,_ := c.String("DEFAULT", "api-host")
    config_url,_ := c.String("DEFAULT", "config-host")
    api_key,_ := c.String("DEFAULT", "api-key")
    fmt.Printf("----%s-%s--\n", config_url, api_key)

    if api_url == "123"  {

    }

    parseJsonFromHttp(config_url, api_key)

    _, err := exec.LookPath("tail")
    if err != nil {
        log.Fatal("installing tail is in your future")
//        exit(1)
    }

 //   filename := "test_logs/test_log.log"
 //   go readData(api_key, api_url, filename)
 //   filename2 := "test_logs/test_log2.log"
//    go readData(api_key, api_url, filename2)

    go getSysStats()

    if err != nil {
        log.Fatal(err)
    }

    err = nil
    for ; err == nil;  {
        //TODO monitor go routines, if one exists reload it
        time.Sleep(0)
        runtime.Gosched()
    }
}
    