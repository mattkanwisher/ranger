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
import "crypto/sha256" 
import "hash"
import "strconv"
import "syscall"

//var BUILD_NUMBER = "_BUILD_"
var BUILD_NUMBER = "1.0.50"
var DOWNLOAD_LOCATION = "http://download.errplane.com/errplane-local-agent-%s"
var OUTPUT_FILE_FORMAT = "errplane-local-agent-%s"
var cmd *exec.Cmd

type AgentConfigType struct {
    Id int
    Version string
    Server string
    Sha256 string
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
func parseJsonFromHttp(api_url string, api_key string) AgentConfigType {
    server := "TEST"
    full_config_url := fmt.Sprintf(api_url + "/api/v1/agents/%s/config?api_key=%s", server, api_key)
    log.Printf("api url %s\n", full_config_url)
    resp, err := http.Get(full_config_url)
    if err != nil {
        // handle error
        fmt.Printf("error getting config data-%s\n",err)    
        os.Exit(1)
    }

    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)

    body2 := []byte(strings.Replace(string(body), "null", "\"\"", -1))//Go doesn't handle nulls in json very well, lets just cheat

    var jsontype  AgentConfigType
    err = json.Unmarshal(body2, &jsontype)
    if err != nil {
        // handle error
        fmt.Printf("error parsing config data-%s\n",err)    
        os.Exit(1)
    }
    return jsontype
}

func write_pid(pid_location string) {
   i := os.Getpid()
   pid := strconv.Itoa(i)
   log.Printf("Writting pid to %s\n", pid_location)
   os.Remove(pid_location)
   ioutil.WriteFile(pid_location, []byte(pid), 0644)
}

func upgrade_version(new_version string, valid_hash string, out_dir string, agent_bin string) {
   log.Printf("Upgrading to current version %s from version %s.\n", new_version, BUILD_NUMBER)

    download_file_url := fmt.Sprintf(DOWNLOAD_LOCATION, new_version)
    log.Printf("download_file %s\n", download_file_url)
    resp, err := http.Get(download_file_url)
    if err != nil {
        // handle error
        fmt.Printf("error getting config data-%s\n",err)    
        os.Exit(1)
    }
    if resp.StatusCode != 200 {
        // handle error
        fmt.Printf("Recieved a bad http code downloading %d-\n", resp.StatusCode)    
        os.Exit(1)
    }

    defer resp.Body.Close()
    download_file, err := ioutil.ReadAll(resp.Body)
    var h hash.Hash = sha256.New()
    h.Write(download_file)
    hash_code := fmt.Sprintf("%x", h.Sum([]byte{}))
    fmt.Printf("downloaded file with hash of %s\n", hash_code)

    if( hash_code == valid_hash) {
        fmt.Printf("Sweet valid file downloaded!")
    } else {
        fmt.Printf("invalid hash!")
        os.Exit(1)
    }

    out_file := fmt.Sprintf(OUTPUT_FILE_FORMAT, new_version)
    out_location := out_dir + out_file


    err = ioutil.WriteFile(out_location, download_file, 0744)
    if err != nil { panic(err) }

    fmt.Printf("Finished writing file!\n")

    //ignore errors
    os.Remove(agent_bin)

    fmt.Printf("symlinking %s to %s\n", out_location, agent_bin)
    err = os.Symlink(out_location, agent_bin)
    if err != nil {
        fmt.Printf("Failed symlinking!--%s\n", err)
        panic(err)
    } 
//Not entirely sure how to use filemode
//    err = os.Chmod(agent_bin, FileMode.)
    cmd = exec.Command("chmod", "+x", agent_bin)
    err = cmd.Start()
    if err != nil {
        fmt.Printf("Failed chmoding!--%s\n", err)
        panic(err)
    } 

    fmt.Printf("Trying new version !\n")
//    agent_bin  = "/Users/kanwisher/projects/errplane/local_agent/local_agent"
//    cmd = exec.Command(agent_bin, "-c", "/Users/kanwisher/projects/errplane/local_agent/config/prod_errplane2.conf" )
//    err = cmd.Start()
    argv := []string {"local_agent"} //, "-c", "/Users/kanwisher/projects/errplane/local_agent/config/prod_errplane2.conf"}
    var proca syscall.ProcAttr
    proca.Env = os.Environ()
    proca.Files =  []uintptr{uintptr(syscall.Stdout), uintptr(syscall.Stderr)}
     _, err = syscall.ForkExec(agent_bin, argv, &proca)//agent_bin)
//     err = syscall.Exec("/Users/kanwisher/projects/errplane/local_agent/local_agent", argv, os.Environ())//agent_bin)
     if err != nil {
        fmt.Printf("Failed running new version!--%s\n", err)
        panic(err)
    } else {
        time.Sleep(10 * time.Second)
        fmt.Printf("Upgraded! Now Extiing! \n")
        os.Exit(0)
    }


}

func checkForUpdatedConfigs(auto_update string, config_url string, api_key string, output_dir string, agent_bin string) {
    for ;;  {
        time.Sleep(10 * time.Second)
        fmt.Printf("Waking up to check configuration\n")

        config_data := parseJsonFromHttp(config_url, api_key)

        log.Printf("Expected agent version-%s\n", config_data.Version)

        if auto_update == "true" && config_data.Version != BUILD_NUMBER {
            upgrade_version(config_data.Version, config_data.Sha256, output_dir, agent_bin)
            log.Printf("Failed upgrading!\n")
        } else {
            log.Printf("Don't need to upgrade versions\n")
        }
    }

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
        fconfig_file = "/etc/errplane.conf"
    }

    log.Print("Loading config file ", *config_file, ".")

    c, _ := config.ReadDefault(fconfig_file)
//    api_url,_ := c.String("DEFAULT", "api-host")
    config_url,_ := c.String("DEFAULT", "config_host")
    api_key,_ := c.String("DEFAULT", "api_key")
    output_dir,_ := c.String("DEFAULT", "agent_path")
    if(len(output_dir) < 1) {
       output_dir =  "/usr/local/errplane/"
    }
    agent_bin,_ := c.String("DEFAULT", "agent_bin")
    if(len(agent_bin) < 1) {
       agent_bin =  "/usr/local/bin/errplane-local-agent"
    }
    pid_location,_ := c.String("DEFAULT", "pid_file")
    if(len(pid_location) < 1) {
        pid_location =  "/var/run/errplane/errplane.pid"
    }
    auto_update,_ := c.String("DEFAULT", "auto_upgrade")

    write_pid(pid_location)

    config_data := parseJsonFromHttp(config_url, api_key)

    log.Printf("Expected agent version-%s\n", config_data.Version)

    if auto_update == "true" && config_data.Version != BUILD_NUMBER {
        upgrade_version(config_data.Version, config_data.Sha256, output_dir, agent_bin)
        os.Exit(1)
    } else {
        log.Printf("Don't need to upgrade versions\n")
    }



    _, err := exec.LookPath("tail")
    if err != nil {
        log.Fatal("installing tail is in your future")
//        exit(1)
    }

    go checkForUpdatedConfigs(auto_update, config_url, api_key, output_dir, agent_bin)

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
    