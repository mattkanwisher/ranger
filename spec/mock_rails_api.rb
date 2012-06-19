require 'rubygems'
require 'sinatra/base'

# this is a mock server for the scala API to hit during testing. Since the scala API only does GETs against the 
# Rails app, this server is setup to echo back stuff that was previous PUT into it.
class MockRailsApi < Sinatra::Base
  configure do
    set :mocks, {}
  end

post '/api/v1/logs/:log_id/agent/:agent_id' do
   puts "Got request app name : log name #{params[:log_id]} -  server #{params[:agent_id]}"
   raw = request.env["rack.input"].read
   puts raw
   "-------"
end

get '/json_api' do
   IO.read("samples/sample_config.json")
end

=begin
  put("/*") do |path|
    settings.mocks[path] = request.body.read
    ""
  end

  get ("/*") do |path|
    if settings.mocks.has_key?(path)
      settings.mocks[path]
    else
      halt 404
    end
  end

  delete("/*") do |path|
    settings.mocks.delete(path)
    ""
  end
=end

end
