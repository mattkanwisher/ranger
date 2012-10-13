package stats

import l4g "code.google.com/p/log4go"
import "fmt"
import "os/exec"
import "strings"
import "strconv"
import "log"
import "io/ioutil"

func parseDriveStats(data string) []EStat {
	var out []EStat
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		if len(line) > 1 {
			sections := strings.Split(line, " ")
			if len(sections) == 2 {
				x := sections[1]
				y, _ := strconv.ParseInt(strings.Replace(sections[0], "%", "", -1), 10, 8)
				out = append(out, EStat{"disk", x, y})

			} else {
				l4g.Debug("Invalid number of sections %d\n", len(sections))
			}
		}
	}
	return out
}

func getDfOutPut() []EStat {
	// var cmdstr string
	//      cmd = exec.Command("df -m  |  sed 's/[  ][   ]*/\\\t/' |  cut  -f 2,3,4,6 | tr -s [:space:] | tail -n+2")
	// cmdstr = "-c \"df -m | tail -n+2 |  sed 's/[ ]/-/'| tr -s [:space:] | cut -d' ' -f2,3,4,5,6 \""
	cmd := exec.Command("bash", "-c", "df -h | tail -n+2 |  sed 's/[ ]/-/'| tr -s [:space:] | rev | cut -d' ' -f1,2 | rev")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	err = nil
	contents, _ := ioutil.ReadAll(stdout)

	scontents := fmt.Sprintf("%s", contents)
	l4g.Debug("df output -%s\n\n", contents)
	parsed := parseDriveStats(scontents)
	l4g.Debug("df parsed output -%s\n\n", parsed)
	return parsed
}
