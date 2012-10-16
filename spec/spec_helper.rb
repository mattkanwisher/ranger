require "rubygems"
require "#{File.dirname(__FILE__)}/j_unit.rb"
require "net/http"
require "yajl"
require "yajl/json_gem"
require "uri"
require "#{File.dirname(__FILE__)}/mock_rails_api.rb"

server = "localhost"
port   = "8086"

RSpec.configure do |config|
  config.before(:all) do
    @pid = fork do
      MockRailsApi.run!
    end
    sleep 2 # wait for the server to start

    @pid_agent = fork do
      File.delete "/tmp/ranger.pid" rescue 0
      ppath = ENV["spec_path"]
      ppath ||= "."
      `GOPATH=\`pwd\` go build -o spec_agent -v  main`
       `#{ppath}/spec_agent -c config/spec_ranger.conf`
    end

    sleep 2 # wait for the server to start
  end

  config.after(:all) do
    `kill -9 #{@pid}`
    `kill -9 #{@pid_agent}`
  end
end

HTTP_CONNECTION = Net::HTTP.new(server, port)

# HTTP helper methods to test the APIs. Each of these should return the response code, headers, and body
def get(path)
  request = Net::HTTP::Get.new(path)
  process_request(request)
end

def get_url(url)
  uri = URI.parse(url)
  request = Net::HTTP::Get.new(uri.path)
  process_request(request, Net::HTTP.new(uri.host, uri.port))
end

def put_url(url, body)
  uri = URI.parse(url)
  request = Net::HTTP::Put.new(uri.path)
  request.body = body
  process_request(request, Net::HTTP.new(uri.host, uri.port))
end

def put(path, body)
  request = Net::HTTP::Put.new(path)
  request.body = body
  process_request(request)
end

def post_with_body(path, body)
  request = Net::HTTP::Post.new(path)
  request.body = body
  process_request(request)
end

def delete(path)
  request = Net::HTTP::Delete.new(path)
  process_request(request)
end

def process_request(request, connection = HTTP_CONNECTION)
  response = connection.request(request)
  header = response.read_header
  headers = {}

  header.each_header do |h|
    headers[h] = header[h]
  end

  {
    :code    => response.code.to_i,
    :body    => response.read_body,
    :headers => headers
  }
end

def to_json(object)
  Yajl::Encoder.encode(object)
end

def parse_json(string)
  Yajl::Parser.parse(string)
end

def puts_response(response)
  json = parse_json(response[:body])
  if json
    puts "\n\n"
    puts json["message"] if json.has_key?("message")
    puts json["backtrace"].join("\n") if json.has_key?("backtrace")
    puts "\n"
    puts response.inspect
  end
end
