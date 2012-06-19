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
