require 'sinatra'

post '/api/v1/logs/:log_id/agent/:agent_id' do
   puts "Got request app name : log name #{params[:log_id]} -  server #{params[:agent_id]}"
   raw = request.env["rack.input"].read
   puts raw
   "-------"
end

#post '/api/v1/applications/test_app/logs/log_filename/servers/bobs_server' do
#  puts "whatever"
#  "Hello World"
#end
