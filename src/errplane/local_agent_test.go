package errplane

import  "testing" //import go package for testing related functionality

func Test_Pid(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T
    pid_file := "/tmp/unit_test.pid"

    write_pid(pid_file)
    btest, _ := FileExist(pid_file)
    if !btest {
        t.Error("Test pid  did not work as expected.") // log error if it did not work as expected
    } else  {
        t.Log("one test passed.") // log some info if you want
    }

}



func Test_FileSystemParse(t *testing.T) {
// df -m
  dfm := `Filesystem           1M-blocks      Used Available Use% Mounted on
/dev/xvda1                8064      6402      1253  84% /
udev                       849         1       849   1% /dev
tmpfs                      342         1       342   1% /run
none                         5         0         5   0% /run/lock
none                       854         0       854   0% /run/shm
/dev/xvda2              342668       328    324934   1% /mnt
`
  res := parseDriveStats(dfm)
  if(res[0].Name == "log_id") {
   t.Log("one test passed.")
  } else {
    t.Error("Test df parse  did not work as expected.") 
  } 
}