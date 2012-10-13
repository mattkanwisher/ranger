package main

import "errplane"
import "errplane/config"

//This next line is auto generated
var BUILD_NUMBER = "_BUILD_"
//var BUILD_NUMBER = "1.0.50"
var DOWNLOAD_LOCATION = "http://download.errplane.com/errplane-local-agent-%s-%s"
var OUTPUT_FILE_FORMAT = "errplane-local-agent-%s"


func main() {
  var defaults config.Consts;
  defaults.Current_build = BUILD_NUMBER;
  defaults.Download_file_location = DOWNLOAD_LOCATION;
  defaults.Output_file_formatter = OUTPUT_FILE_FORMAT;
  errplane.Errplane_main(defaults)
}