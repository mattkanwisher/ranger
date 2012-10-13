package logsp

import "os/exec"
import "log"
import "bufio"

import "errplane/config"

func ReadLogData(filename string, log_id int, logOutputChan chan<- *config.LogTuple, deathChan chan<- *string) {
	cmd := exec.Command("tail", "-f", filename)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	err = nil
	reader := bufio.NewReader(stdout)
	sbuffer := ""
	lines := 0
	for err == nil {
		s, err := reader.ReadString('\n')
		sbuffer += s
		lines += 1
		if lines > 5 { //|| time > 1 min) {
			logOutputChan <- &config.LogTuple{log_id, sbuffer}
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
