package config
import l4g "code.google.com/p/log4go"
import "time"
import "runtime"
import "os"
import "fmt"
import "net/url"
import "net/http"
import "strings"
import "encoding/json"
import "io/ioutil"

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

type Consts struct {
    Output_file_formatter string;
    Download_file_location string;
    Current_build string;
}

//TODO: IF this fails  read from disk, if that fails sleep until the server is back online
func ParseJsonFromHttp(api_url string, api_key string) AgentConfigType {
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



func CheckForUpdatedConfigs(auto_update string, config_url string, api_key string, output_dir string, agent_bin string, out chan<- *AgentConfigType, defaults Consts) {
    for ;;  {
        config_data := ParseJsonFromHttp(config_url, api_key)
        out <- &config_data


        if auto_update == "true" && config_data.Version != defaults.Current_build {
          l4g.Debug("Upgrading agent version-%s\n", config_data.Version)
          hdata := config_data.Sha256
          if( runtime.GOARCH == "amd64" ) {
             hdata = config_data.Sha256_amd64
          }
          Upgrade_version(config_data.Version, hdata, output_dir, agent_bin, defaults)
          l4g.Debug("Failed upgrading!\n")
        } else {
            l4g.Fine("Don't need to upgrade versions\n")
        }
        time.Sleep(10 * time.Second)
    }

}