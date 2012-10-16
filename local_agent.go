package main

import "ranger"
import "ranger/config"

//This next line is auto generated
var BUILD_NUMBER = "_BUILD_"

//var BUILD_NUMBER = "1.0.50"
var DOWNLOAD_LOCATION = "http://download.ranger.com/ranger-local-agent-%s-%s"
var OUTPUT_FILE_FORMAT = "ranger-local-agent-%s"

func main() {
	var defaults config.Consts
	defaults.Current_build = BUILD_NUMBER
	defaults.Download_file_location = DOWNLOAD_LOCATION
	defaults.Output_file_formatter = OUTPUT_FILE_FORMAT
	ranger.Errplane_main(defaults)
}
