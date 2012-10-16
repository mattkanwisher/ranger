package ranger

import "fmt"
import "net/http"
import "bytes"
import "log"
import "os/exec"
import "time"
import "runtime"
import "github.com/kless/goconfig/config"
import "os"
import "io/ioutil"
import "github.com/droundy/goopt"
import "strconv"
import "syscall"
import "net/url"
import l4g "code.google.com/p/log4go"
import "ranger/stats"
import "ranger/logsp"
import econfig "ranger/config"
import "ranger/daemonize"

var cmd *exec.Cmd

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

// takes a base64 encoded string that can be associated with a data point in the format: <name> <value> <time in seconds> <base64 string>" do

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
func write_pid(pid_location string) {
	i := os.Getpid()
	pid := strconv.Itoa(i)
	l4g.Debug("Writting pid to %s\n", pid_location)
	os.Remove(pid_location)
	ioutil.WriteFile(pid_location, []byte(pid), 0644)
}

//TODO get the poll interval from the brain
func dataPosting(logOutputChan <-chan *econfig.LogTuple, statOutputChan <-chan *[]stats.EStat, api_key string, api_url string) {
	statusInterval := 1 * time.Second //Default it and let the brain update it later
	ticker := time.NewTicker(statusInterval)
	buffer := make(map[int]string)

	for {
		select {
		case <-ticker.C:
			for lid, data := range buffer {
				if len(data) > 0 {
					postData(api_key, api_url, data, lid)
					buffer[lid] = ""
				}
			}
		case log_tup := <-logOutputChan:
			buffer[log_tup.Log_id] += log_tup.Data
		case estats := <-statOutputChan:
			output := ""
			for _, data := range *estats {
				fmt.Printf("WOOOT GOT A STAT %s %d \n", data.Name, data.Val)
				output += fmt.Sprintf("%s %d %d \n", data.Name, data.Val, time.Now().Unix())
			}
			stats.PostStatData(api_key, api_url, output)
		}
	}
}

func theBrain(in <-chan *econfig.AgentConfigType, api_key string, api_url string) {
	runningGoR := make(map[string]bool)

	logOutputChan := make(chan *econfig.LogTuple)
	gorDeathChan := make(chan *string)
	statOutputChan := make(chan *[]stats.EStat)

	//Setup go routine for Data posting
	go dataPosting(logOutputChan, statOutputChan, api_key, api_url)
	runningGoR["SYSTEM_DATA_POST"] = true

	//TODO for now always run system stats go routine
	go stats.GetSysStats(statOutputChan)
	runningGoR["SYSTEM_STATS"] = true

	for {
		l4g.Debug("Waiting for config data")
		//TODO LOOK FOR DEATH OF GOROUTINES AND RESPAWN THEM
		for {
			select {
			case config_data := <-in:
				l4g.Debug("Recieved for config data")
				for _, alog := range config_data.Agent_logs {
					if runningGoR[alog.Log.Path] == false {
						go logsp.ReadLogData(alog.Log.Path, alog.Log.Id, logOutputChan, gorDeathChan)
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
		r1 = 0
	}

	return r1, 0

}

var config_file = goopt.String([]string{"-c", "--config"}, "/etc/ranger.conf", "config file")
var install_api_key = goopt.String([]string{"-i", "--install-api-key"}, "", "install api key")
var amForeground = goopt.Flag([]string{"--foreground"}, []string{"--background"}, "run foreground", "run background")

func setup_logger() {
	filename := "/var/log/ranger/ranger.log"

	// Create a default logger that is logging messages of FINE or higher to filename, no rotation
	//    log.AddFilter("file", l4g.FINE, l4g.NewFileLogWriter(filename, false))

	// =OR= Can also specify manually via the following: (these are the defaults, this is equivalent to above)
	flw := l4g.NewFileLogWriter(filename, false)
	if flw == nil {
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

func Errplane_main(defaults econfig.Consts) {
	fmt.Printf("ERRPlane Local Agent starting, Version %s \n", defaults.Current_build)

	goopt.Description = func() string {
		return "ERRPlane Local Agent."
	}
	goopt.Version = defaults.Current_build
	goopt.Summary = "ErrPlane Log and System Monitor"
	goopt.Parse(nil)

	var fconfig_file string
	fconfig_file = *config_file

	fmt.Printf("Loading config file %s.\n", fconfig_file)

	c, err := config.ReadDefault(fconfig_file)
	if err != nil {
		log.Fatal("Can not find the Errplane Config file, please install it in /etc/ranger/ranger.conf.")
	}

	api_key, _ := c.String("DEFAULT", "api_key")

	if len(*install_api_key) > 1 {
		fmt.Printf("Saving new Config!\n")
		c.AddOption("DEFAULT", "api_key", *install_api_key)
		c.WriteFile(fconfig_file, 0644, "")

		c, err := config.ReadDefault(fconfig_file)
		if err != nil {
			log.Fatal("Can not find the Errplane Config file, please install it in /etc/ranger/ranger.conf.")
		}
		api_key, _ = c.String("DEFAULT", "api_key")
	}

	if len(api_key) < 1 {
		log.Fatal("No api key found. Please rerun this with --install-api-key <api_key_here> ")
		os.Exit(1)
	}

	if !*amForeground {
		//Daemonizing requires root
		if os.Getuid() == 0 {
			daemonize.Do_fork(os.Args[0], fconfig_file)
			l4g.Fine("Exiting parent process-%s\n", os.Args[0])
			os.Exit(0)
		} else {
			fmt.Printf("Daemoning requires root \n")
			os.Exit(1)
		}
	}

	setup_logger()

	api_url, _ := c.String("DEFAULT", "api_host")
	config_url, _ := c.String("DEFAULT", "config_host")
	output_dir, _ := c.String("DEFAULT", "agent_path")
	if len(output_dir) < 1 {
		output_dir = "/usr/local/ranger/"
	}
	agent_bin, _ := c.String("DEFAULT", "agent_bin")
	if len(agent_bin) < 1 {
		agent_bin = "/usr/local/bin/ranger-local-agent"
	}
	pid_location, _ := c.String("DEFAULT", "pid_file")
	if len(pid_location) < 1 {
		pid_location = "/var/run/ranger/ranger.pid"
	}
	auto_update, _ := c.String("DEFAULT", "auto_upgrade")

	write_pid(pid_location)

	config_data := econfig.ParseJsonFromHttp(config_url, api_key)

	l4g.Debug("Expected agent version-%s\n", config_data.Version)

	if auto_update == "true" && config_data.Version != defaults.Current_build {
		econfig.Upgrade_version(config_data.Version, config_data.Sha256, output_dir, agent_bin, defaults)
		os.Exit(1)
	} else {
		l4g.Debug("Don't need to upgrade versions\n")
	}

	_, err = exec.LookPath("tail")
	if err != nil {
		log.Fatal("installing tail is in your future")
		//        exit(1)
	}

	configChan := make(chan *econfig.AgentConfigType)

	go theBrain(configChan, api_key, api_url)

	go econfig.CheckForUpdatedConfigs(auto_update, config_url, api_key, output_dir, agent_bin, configChan, defaults)

	if err != nil {
		log.Fatal(err)
	}

	err = nil
	for err == nil {
		//TODO monitor go routines, if one exists reload it
		time.Sleep(0)
		runtime.Gosched()
	}
}
