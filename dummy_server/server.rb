require 'sinatra'

post '/hi' do
   raw = request.env["rack.input"].read
   puts raw
  "Hello World!"
end
