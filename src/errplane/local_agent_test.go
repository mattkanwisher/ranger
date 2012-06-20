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
  /*
  dfm := `380516 301244 79021 79% /
0 0 0 100% /dev
95300 47736 47563 50% /Volumes/BOOTCAMP
0 0 0 100% /net
0 0 0 100% /home
8063 6439 1213 84% /Volumes/lcsdomain.eikonswift.com`
*/
  dfm := `8064 6440 1215 85% /
829 1 829 1% /dev
334 1 334 1% /run
5 0 5 0% /run/lock
834 0 834 0% /run/shm
150293 188 142471 1% /mnt`

  res := parseDriveStats(dfm)
  t.Log("got data back %x\n", res)
  if(res[0].Group == "/") && (res[0].Name == "1M-blocks") && (res[0].Val == 127) {
   t.Log("one test passed.")
  } else {
    t.Log(res[0])
    t.Error("Test df parse  did not work as expected." ) 
  } 

  if(res[1].Group == "/") && (res[1].Name == "AvailablePer") && (res[1].Val == 85) {
   t.Log("one test passed.")
  } else {
    t.Error("Test df parse  did not work as expected.") 
  } 

}
/*
func Test_TopParseLinux(t *testing.T) {
// df -m
  dfm := `8064 6440 1215 85% /
829 1 829 1% /dev
334 1 334 1% /run
5 0 5 0% /run/lock
834 0 834 0% /run/shm
150293 188 142471 1% /mnt`
  res := parseTopStatsLinux(dfm)
  if(res[0].Group == "/dev/xvda1") && (res[0].Name == "Filesystem") {
   t.Log("one test passed.")
  } else {
    t.Error("Test df parse  did not work as expected.") 
  } 
}
*/