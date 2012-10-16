package daemonize

import . "ranger/common"
import "fmt"
import "os"
import "os/exec"
import "log"

var cmd *exec.Cmd

func Do_fork(executable, config_file string) {
	executable = "/etc/init.d/ranger"

	bfile, _ := FileExist(executable)
	if bfile == false {
		fmt.Printf("Can not find daemonizing script " + executable + "! \n")
		os.Exit(1)
	}

	cmdstr := executable + " -c " + config_file + " --foreground"
	fmt.Printf(cmdstr + "\n")

	cmd = exec.Command(executable, "start")

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
}
