package errplane

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
//import "syscall"
import "net/url"

var BUILD_NUMBER = "_BUILD_"
//var BUILD_NUMBER = "1.0.50"
var DOWNLOAD_LOCATION = "http://download.errplane.com/errplane-local-agent-%s"
var OUTPUT_FILE_FORMAT = "errplane-local-agent-%s"
var cmd *exec.Cmd

func FileExist(name string) (bool, error) { 
    _, err := os.Stat(name) 
    if err == nil { 
            return true, nil 
    } 
    return false, err 
} 
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

type LogTuple struct { 
   Log_id int; 
   Data string; 
} 

type EStat struct {
    Group string;
    Name string;
    Val int64;
}


func postData(api_key string, api_url string, data string, log_id int) {
    log.Printf("Posting log data to server %s\n", log_id)
    server_name, _ := os.Hostname() 
//    contents,_ := ioutil.ReadAll(data);
//  buf := bytes.NewBuffer("your string")
  buf2 := bytes.NewBufferString(data)
    //TODO url escaping
    url := fmt.Sprintf("%s/api/v1/logs/%d/agents/%s?api_key=%s", api_url, log_id, url.QueryEscape(server_name), api_key)
    log.Printf("posting to url -%s\n%s\n", url,buf2)
    http.Post(url, "application/text", buf2)
//TODO HANDLE ERROR AND RETRIES!
}

func parseDriveStats(data string) []EStat {
   var out  []EStat
   lines := strings.Split(data, "\n")
   for _, line := range lines {
     if(len(line) > 3) {
        sections := strings.Split(line, " ")
        if( len(sections) == 5 ) {
            x, _ := strconv.ParseInt(sections[0], 10, 8)
            out = append(out, EStat{ sections[4],  "1M-blocks", x})
            y, _ := strconv.ParseInt(strings.Replace(sections[3], "%", "",-1),  10, 8)
            out = append(out, EStat{ sections[4],  "AvailablePer", y})

        } else {
            log.Printf("Invalid number of sections %d\n", len(sections))
        }        
     }
   }
   return out
}


func parseTopStatsLinux(data string) []EStat {
    return nil;
}

func getTopOutPut() {
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

func getDfOutPut() {
   // var cmdstr string
//      cmd = exec.Command("df -m  |  sed 's/[  ][   ]*/\\\t/' |  cut  -f 2,3,4,6 | tr -s [:space:] | tail -n+2")
   // cmdstr = "-c \"df -m | tail -n+2 |  sed 's/[ ]/-/'| tr -s [:space:] | cut -d' ' -f2,3,4,5,6 \""
    cmd = exec.Command("bash", "-c", "df -m | tail -n+2 |  sed 's/[ ]/-/'| tr -s [:space:] | cut -d' ' -f2,3,4,5,6 ")

    stdout, err := cmd.StdoutPipe()
    if err != nil {
        log.Fatal(err)
    }
    if err := cmd.Start(); err != nil {
        log.Fatal(err)
    }

    err  = nil
    contents,_ := ioutil.ReadAll(stdout)

    scontents := fmt.Sprintf("%s", contents)
    log.Printf("df output -%s\n\n", contents)    
    parsed := parseDriveStats(scontents)
    log.Printf("df parsed output -%s\n\n", parsed)    
}

func getSysStats() {
    log.Printf("Pulling system stats")

    //getTopOutPut() 
    getDfOutPut()
}

func readLogData(filename string, log_id int, logOutputChan chan<- *LogTuple, deathChan chan<- *string) {
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
        sbuffer += s
        lines += 1
        if(lines > 5 ) { //|| time > 1 min) {
            logOutputChan <- &LogTuple{ log_id, sbuffer}
            sbuffer = ""
            lines = 0
        }
        if err != nil {
            deathChan <- &filename
            return
        }
    }

    //SetFinalizer
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
    server, _ :=  os.Hostname() 
    full_config_url := fmt.Sprintf(api_url + "/api/v1/agents/%s/config?api_key=%s", url.QueryEscape(server), api_key)
    log.Printf("Updating config info %s\n", full_config_url)
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
    //argv := []string {"local_agent"} //, "-c", "/Users/kanwisher/projects/errplane/local_agent/config/prod_errplane2.conf"}
    //var proca syscall.ProcAttr
    //proca.Env = os.Environ()
    //proca.Files =  []uintptr{uintptr(syscall.Stdout), uintptr(syscall.Stderr)}
//     _, err = syscall.ForkExec(agent_bin, argv, &proca)//agent_bin)
//     err = syscall.Exec("/Users/kanwisher/projects/errplane/local_agent/local_agent", argv, os.Environ())//agent_bin)
    //TODO FOR NOW WE JUST EXIT AFTER UPGRADE AND LET MONIT RESTART US, UGH HAVE TO FIGURE OUT THIS FORKEXEC BS
    err = nil
     if err != nil {
        fmt.Printf("Failed running new version!--%s\n", err)
        panic(err)
    } else {
        time.Sleep(10 * time.Second)
        fmt.Printf("Upgraded! Now Extiing! \n")
        os.Exit(0)
    }


}

//TODO get the poll interval from the brain
func dataPosting(logOutputChan <-chan *LogTuple, api_key string, api_url string) {
    statusInterval := 1 * time.Second //Default it and let the brain update it later
    ticker := time.NewTicker(statusInterval)
    buffer := make(map[int]string)
    for {
      select {
      case <-ticker.C:
        for lid,data := range buffer {
            if( len(data) > 0) {
                postData(api_key, api_url, data, lid)
                buffer[lid] = ""
            }
        }
      case log_tup := <-logOutputChan:
          buffer[log_tup.Log_id] += log_tup.Data
      }
    }
}

func theBrain( in <-chan *AgentConfigType, api_key string, api_url string) {
    runningGoR := make(map[string]bool)


    logOutputChan := make(chan *LogTuple)
    gorDeathChan := make(chan *string)

    //Setup go routine for Data posting
    go dataPosting(logOutputChan, api_key, api_url)
    runningGoR["SYSTEM_DATA_POST"] = true

    //TODO for now always run system stats go routine
    go getSysStats()
    runningGoR["SYSTEM_STATS"] = true

    for ;; {
      log.Printf("Waiting for config data")
      //TODO LOOK FOR DEATH OF GOROUTINES AND RESPAWN THEM
      for {
        select {
        case config_data := <-in:
          log.Printf("Recieved for config data")
          for _,alog := range config_data.Agent_logs { 
             if( runningGoR[alog.Log.Path] == false) {
               go readLogData(alog.Log.Path, alog.Log.Id, logOutputChan, gorDeathChan)
               runningGoR[alog.Log.Path] = true
               log.Printf("Launched go routine\n")
              }
          }
        case death := <-gorDeathChan:
            log.Printf("death-%s\n", *death)
            runningGoR[*death] = false
        }
      }
    }
}


func checkForUpdatedConfigs(auto_update string, config_url string, api_key string, output_dir string, agent_bin string, out chan<- *AgentConfigType) {
    for ;;  {
        config_data := parseJsonFromHttp(config_url, api_key)
        out <- &config_data


        if auto_update == "true" && config_data.Version != BUILD_NUMBER {
          log.Printf("Upgrading agent version-%s\n", config_data.Version)
          upgrade_version(config_data.Version, config_data.Sha256, output_dir, agent_bin)
          log.Printf("Failed upgrading!\n")
        } else {
            log.Printf("Don't need to upgrade versions\n")
        }
        time.Sleep(10 * time.Second)
    }

}

var config_file = goopt.String([]string{"-c", "--config"}, "/etc/errplane.conf", "config file")

func Errplane_main() {
    fmt.Printf("ERRPlane Local Agent starting, Version %s \n", BUILD_NUMBER)

    goopt.Description = func() string {
        return "ERRPlane Local Agent."
    }
    goopt.Version = BUILD_NUMBER
    goopt.Summary = "ErrPlane Log and System Monitor"
    goopt.Parse(nil)

    var fconfig_file string
    fconfig_file = *config_file

    fmt.Printf("Loading config file %s.\n", fconfig_file )

    c, err := config.ReadDefault(fconfig_file)
    if( err != nil ) {
      log.Fatal("Can not find the Errplane Config file, please install it in /etc/errplane/errplane.conf.")
    }

    api_url,_ := c.String("DEFAULT", "api_host")
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



    _, err = exec.LookPath("tail")
    if err != nil {
        log.Fatal("installing tail is in your future")
//        exit(1)
    }

    configChan := make(chan *AgentConfigType)

    go theBrain(configChan, api_key, api_url)

    go checkForUpdatedConfigs(auto_update, config_url, api_key, output_dir, agent_bin, configChan)

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
