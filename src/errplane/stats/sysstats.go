package stats
import "runtime"
import "os/exec"
import "log"
import "io/ioutil"
import l4g "code.google.com/p/log4go"

var cmd *exec.Cmd

func parseTopStatsLinux(data string) []EStat {
	return nil
}

func getTopOutPut() {
	myos := runtime.GOOS
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

	err = nil
	contents, _ := ioutil.ReadAll(stdout)

	l4g.Debug("top output -%s\n\n========\n%s\n", contents, myos)
}
