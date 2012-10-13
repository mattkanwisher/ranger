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
import "syscall"
import "net/url"
import l4g "code.google.com/p/log4go"

var BUILD_NUMBER = "_BUILD_"
//var BUILD_NUMBER = "1.0.50"
var DOWNLOAD_LOCATION = "http://download.errplane.com/errplane-local-agent-%s-%s"
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
    Sha256 string //i386
    Sha256_amd64 string
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
    l4g.Debug("Posting log data to server %s\n", log_id)
    server_name, _ := os.Hostname() 
//    contents,_ := ioutil.ReadAll(data);
//  buf := bytes.NewBuffer("your string")
  buf2 := bytes.NewBufferString(data)
    //TODO url escaping
    url := fmt.Sprintf("%s/api/v1/logs/%d/agents/%s?api_key=%s", api_url, log_id, url.QueryEscape(server_name), api_key)
    l4g.Debug("posting to url -%s\n", url)
    http.Post(url, "application/text", buf2)
//TODO HANDLE ERROR AND RETRIES!
}

func postStatData(api_key string, api_url string, name string, value int) {
    l4g.Debug("Posting stat data to server %s\n", name_string)
    server_name, _ := os.Hostname() 
//    contents,_ := ioutil.ReadAll(data);
//  buf := bytes.NewBuffer("your string")
    buf2 := bytes.NewBufferString(data)
    //TODO url escaping
    url := fmt.Sprintf("%s//api/v1/time_series/system_%s_%d?api_key=%s", api_url, url.QueryEscape(server_name), value, api_key)
    l4g.Debug("posting to url -%s\n", url)
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
            l4g.Debug("Invalid number of sections %d\n", len(sections))
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

    l4g.Debug("top output -%s\n\n========\n%s\n", contents, myos)    
}

func getDfOutPut() []EStat {
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
    l4g.Debug("df output -%s\n\n", contents)    
    parsed := parseDriveStats(scontents)
    l4g.Debug("df parsed output -%s\n\n", parsed)    
    return parsed
}

func getSysStats( statOutputChan chan<- *[]EStat) {
    l4g.Debug("Pulling system stats")

    for  {
      //topstat := getTopOutPut() 
      //statOutputChan <- &topstat
      dfstat := getDfOutPut()
      fmt.Printf("putting data to stat channel")
      statOutputChan <- &dfstat

      //TODO listen for a death channel
      time.Sleep(10 * time.Second)
    }
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
    l4g.Debug("Updating config info %s\n", full_config_url)
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
   l4g.Debug("Writting pid to %s\n", pid_location)
   os.Remove(pid_location)
   ioutil.WriteFile(pid_location, []byte(pid), 0644)
}

func upgrade_version(new_version string, valid_hash string, out_dir string, agent_bin string) {
   l4g.Debug("Upgrading to current version %s from version %s.\n", new_version, BUILD_NUMBER)

    download_file_url := fmt.Sprintf(DOWNLOAD_LOCATION, new_version, runtime.GOARCH)
    l4g.Debug("download_file %s\n", download_file_url)
    resp, err := http.Get(download_file_url)
    if err != nil {
        // handle error
        l4g.Error("error getting config data-%s\n",err)    
        return
    }
    if resp.StatusCode != 200 {
        // handle error
        l4g.Error("Recieved a bad http code downloading %d-\n", resp.StatusCode)    
        return
    }

    defer resp.Body.Close()
    download_file, err := ioutil.ReadAll(resp.Body)
    var h hash.Hash = sha256.New()
    h.Write(download_file)
    hash_code := fmt.Sprintf("%x", h.Sum([]byte{}))
    fmt.Printf("downloaded file with hash of %s\n", hash_code)

    if( hash_code == valid_hash) {
        l4g.Debug("Sweet valid file downloaded!")
    } else {
        l4g.Error("invalid hash!")
        return
    }

    out_file := fmt.Sprintf(OUTPUT_FILE_FORMAT, new_version)
    out_location := out_dir + out_file


    err = ioutil.WriteFile(out_location, download_file, 0744)
    if err != nil {
       l4g.Error(err) 
       return
     }

    fmt.Printf("Finished writing file!\n")

    //ignore errors
    os.Remove(agent_bin)

    fmt.Printf("symlinking %s to %s\n", out_location, agent_bin)
    err = os.Symlink(out_location, agent_bin)
    if err != nil {
        l4g.Error("Failed symlinking!--%s\n", err)
        return
    } 
//Not entirely sure how to use filemode
//    err = os.Chmod(agent_bin, FileMode.)
    cmd = exec.Command("chmod", "+x", agent_bin)
    err = cmd.Start()
    if err != nil {
        l4g.Error("Failed chmoding!--%s\n", err)
        return
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
    //TODO for now just launch the daemon script instead of some insane fork/exec stufff
    if err != nil {
        l4g.Error("Failed running new version!--%s\n", err)
        return
    } else {
        l4g.Debug("Upgrading! Please wait! \n")
        do_fork("","")
        time.Sleep(10 * time.Second)
        l4g.Debug("Upgraded! Now Extiing! \n")
        os.Exit(0)
    }


}

//TODO get the poll interval from the brain
func dataPosting(logOutputChan <-chan *LogTuple, statOutputChan <-chan *[]EStat, api_key string, api_url string) {
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
      case estats := <-statOutputChan:
          for _,data := range *estats {
          fmt.Printf("WOOOT GOT A STAT %s %d \n", data.Name, data.Val)

          }
      }
    }
}

func theBrain( in <-chan *AgentConfigType, api_key string, api_url string) {
    runningGoR := make(map[string]bool)


    logOutputChan := make(chan *LogTuple)
    gorDeathChan := make(chan *string)
    statOutputChan := make(chan *[]EStat)

    //Setup go routine for Data posting
    go dataPosting(logOutputChan, statOutputChan, api_key, api_url)
    runningGoR["SYSTEM_DATA_POST"] = true

    //TODO for now always run system stats go routine
    go getSysStats(statOutputChan)
    runningGoR["SYSTEM_STATS"] = true

    for ;; {
      l4g.Debug("Waiting for config data")
      //TODO LOOK FOR DEATH OF GOROUTINES AND RESPAWN THEM
      for {
        select {
        case config_data := <-in:
          l4g.Debug("Recieved for config data")
          for _,alog := range config_data.Agent_logs { 
             if( runningGoR[alog.Log.Path] == false) {
               go readLogData(alog.Log.Path, alog.Log.Id, logOutputChan, gorDeathChan)
               runningGoR[alog.Log.Path] = true
               l4g.Debug("Launched go routine\n")
              }
          }
        case death := <-gorDeathChan:
            l4g.Debug("death-%s\n", *death)
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
          l4g.Debug("Upgrading agent version-%s\n", config_data.Version)
          hdata := config_data.Sha256
          if( runtime.GOARCH == "amd64" ) {
             hdata = config_data.Sha256_amd64
          }
          upgrade_version(config_data.Version, hdata, output_dir, agent_bin)
          l4g.Debug("Failed upgrading!\n")
        } else {
            l4g.Fine("Don't need to upgrade versions\n")
        }
        time.Sleep(10 * time.Second)
    }

}
/*
func daemon (nochdir, noclose int) int {
  var ret uintptr
  var err uintptr

  ret,_,err = syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
  if err != 0 { return -1 }
  switch (ret) {
    case 0:
      break
    default:
      os.Exit (0)
  }

  if syscall.Setsid () == -1 { return -1 }
  if (nochdir == 0) { os.Chdir("/") }

  if noclose == 0 {
    f, e := os.Open ("/dev/null", os.O_RDWR, 0)
    if e == nil {
      fd := f.Fd ()
      syscall.Dup2 (fd, os.Stdin.Fd ())
      syscall.Dup2 (fd, os.Stdout.Fd ())
      syscall.Dup2 (fd, os.Stderr.Fd ())
    }
  }

  return 0
}
*/


func fork() (pid uintptr, err syscall.Errno) {
    var r1, r2 uintptr
    var err1 syscall.Errno

    darwin := runtime.GOOS == "darwin"

    r1, r2, err1 = syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)

    if err != 0 {
        return 0, err1
    }

    // Handle exception for darwin
    if darwin && r2 == 1 {
        r1 = 0;
    }

    return r1, 0

}


func do_fork(executable, config_file string) {
  executable = "/etc/init.d/errplane"
  
  bfile,_ := FileExist(executable)
  if( bfile == false ){
    fmt.Printf("Can not find daemonizing script " + executable + "! \n")
    os.Exit(1)
  }

  cmdstr :=   executable + " -c " + config_file + " --foreground"
  fmt.Printf(cmdstr + "\n")

  cmd = exec.Command( executable, "start" )

  if err := cmd.Start(); err != nil {
      log.Fatal(err)
  }
}

var config_file = goopt.String([]string{"-c", "--config"}, "/etc/errplane.conf", "config file")
var install_api_key = goopt.String([]string{"-i", "--install-api-key"},  "", "install api key")
var amForeground = goopt.Flag([]string{"--foreground"}, []string{"--background"},  "run foreground", "run background")

func setup_logger() {
  filename := "/var/log/errplane/errplane.log"

  // Create a default logger that is logging messages of FINE or higher to filename, no rotation
//    log.AddFilter("file", l4g.FINE, l4g.NewFileLogWriter(filename, false))

  // =OR= Can also specify manually via the following: (these are the defaults, this is equivalent to above)
  flw := l4g.NewFileLogWriter(filename, false)
  if(flw == nil){
    fmt.Printf("No permission to write to %s, going to switch to stdout only\n", filename)
  } else {
    flw.SetFormat("[%D %T] [%L] (%S) %M")
    flw.SetRotate(false)
    flw.SetRotateSize(0)
    flw.SetRotateLines(0)
    flw.SetRotateDaily(true)
    l4g.AddFilter("file", l4g.DEBUG, flw)

    l4g.AddFilter("stdout", l4g.ERROR, l4g.NewConsoleLogWriter())
  }
}

func test_read( quit <-chan *string) {

  cmd := exec.Command("tail", "-f", "/tmp/matt.txt")
  stdout, err := cmd.StdoutPipe()
  if err != nil {
      log.Fatal(err)
  }
  if err := cmd.Start(); err != nil {
      log.Fatal(err)
  }

  err  = nil
  
  buff := make([]byte,1024)
  for {
      n,_ := stdout.Read(buff)
      if n < 1  {
        fmt.Printf(" ")
        break
      } else {
        fmt.Printf("0")
        fmt.Sprintf("found data -%s", n)
      }
  }
  /*
  reader := bufio.NewReader(stdout)

    for ; err == nil;  {
        s,err := reader.ReadString('\n')
        if(err != nil){
          log.Fatal(err)
        }
        fmt.Printf("-%s@", s)
      }
      */
//    outChan <- buff
  
}

func Errplane_main() {
    fmt.Printf("ERRPlane Local Agent starting, Version %s \n", BUILD_NUMBER)
    quitChan := make(chan *string)

    test_read(quitChan)

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

    api_key,_ := c.String("DEFAULT", "api_key")

    if(len(*install_api_key) > 1) {
      fmt.Printf("Saving new Config!\n")
      c.AddOption("DEFAULT", "api_key", *install_api_key)
      c.WriteFile(fconfig_file, 0644, "")

      c, err := config.ReadDefault(fconfig_file)
      if( err != nil ) {
        log.Fatal("Can not find the Errplane Config file, please install it in /etc/errplane/errplane.conf.")
      }
      api_key,_ = c.String("DEFAULT", "api_key")
    }

    if(len(api_key) < 1) {
      log.Fatal("No api key found. Please rerun this with --install-api-key <api_key_here> ")      
      os.Exit(1)
    }

    if(!*amForeground) {
      //Daemonizing requires root
      if( os.Getuid() == 0 ) {
        do_fork(os.Args[0], fconfig_file)
        l4g.Fine("Exiting parent process-%s\n", os.Args[0])
        os.Exit(0)
        } else {
          fmt.Printf("Daemoning requires root \n")
          os.Exit(1)
        }
    }

    setup_logger()

    api_url,_ := c.String("DEFAULT", "api_host")
    config_url,_ := c.String("DEFAULT", "config_host")
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

    l4g.Debug("Expected agent version-%s\n", config_data.Version)

    if auto_update == "true" && config_data.Version != BUILD_NUMBER {
        upgrade_version(config_data.Version, config_data.Sha256, output_dir, agent_bin)
        os.Exit(1)
    } else {
        l4g.Debug("Don't need to upgrade versions\n")
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
