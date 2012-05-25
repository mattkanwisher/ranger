require 'sinatra'

post '/api/v1/applications/:app_name/logs/:log_name/servers/:server_name' do
   puts "Got request app name : #{params[:app_name]} - log name #{params[:log_name]} -  server #{params[:server_name]}"
   raw = request.env["rack.input"].read
   puts raw
   "-------"
end

#post '/api/v1/applications/test_app/logs/log_filename/servers/bobs_server' do
#  puts "whatever"
#  "Hello World"
#end
