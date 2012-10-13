package config
import l4g "code.google.com/p/log4go"
import "fmt"
import "runtime"
import "crypto/sha256" 
import "net/http"
import "os"
import "io/ioutil"
import "os/exec"
import "hash"
import "time"
import "errplane/daemonize"


var cmd *exec.Cmd

func Upgrade_version(new_version string, valid_hash string, out_dir string, agent_bin string, defaults Consts ) {
   l4g.Debug("Upgrading to current version %s from version %s.\n", new_version, defaults.Current_build)

    download_file_url := fmt.Sprintf(defaults.Download_file_location, new_version, runtime.GOARCH)
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

    out_file := fmt.Sprintf(defaults.Output_file_formatter, new_version)
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
        daemonize.Do_fork("","")
        time.Sleep(10 * time.Second)
        l4g.Debug("Upgraded! Now Extiing! \n")
        os.Exit(0)
    }


}