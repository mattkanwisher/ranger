package stats

import l4g "code.google.com/p/log4go"
import "time"
import "fmt"
import "os"
import "bytes"
import "net/url"
import "net/http"

type EStat struct {
	Group string
	Name  string
	Val   int64
}

func GetSysStats(statOutputChan chan<- *[]EStat) {
	l4g.Debug("Pulling system stats")

	for {
		//topstat := getTopOutPut() 
		//statOutputChan <- &topstat
		dfstat := getDfOutPut()
		fmt.Printf("putting data to stat channel")
		statOutputChan <- &dfstat

		//TODO listen for a death channel
		time.Sleep(10 * time.Second)
	}
}

func PostStatData(api_key string, api_url string, data string) {
	server_name, _ := os.Hostname()
	l4g.Debug("Posting stat data to server %s\n", data)
	//    contents,_ := ioutil.ReadAll(data);
	//  buf := bytes.NewBuffer("your string")
	buf2 := bytes.NewBufferString(data)
	//TODO url escaping
	url := fmt.Sprintf("%s/api/v2/time_series/applications/system_%s/environments/production?api_key=%s", api_url, url.QueryEscape(server_name), api_key)
	l4g.Debug("posting to url -%s\n", url)
	http.Post(url, "application/text", buf2)
	//TODO HANDLE ERROR AND RETRIES!
}
