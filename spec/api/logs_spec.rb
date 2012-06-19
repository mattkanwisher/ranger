require "#{File.dirname(__FILE__)}/../spec_helper.rb"

describe "the log file api" do
  before(:all) do
    @organization = {:id => 1}
    put_url("http://localhost:4567/api/v1/organizations", @organization.to_json)

    @log_id = "1_#{Time.now.to_i}"
    @agent_id = "10.0.3.23"

    @body1 = "line 1\nline 2\nline 3"
    @body2 = "another line\nand another"
  end

  it "posts multiple lines from a log file (oldest first line to newest last line) with the current time" do
    response = post_with_body("/api/v1/logs/#{@log_id}/agents/#{@agent_id}?api_key=ignored", @body1)
    response[:code].should == 200
    
    response = get("/api/v1/logs/#{@log_id}/agents/#{@agent_id}?api_key=ignored")
    lines =  parse_json(response[:body])
    ids = lines.map {|l| l[0]}
    times = lines.map {|l| l[1]}
    content = lines.map {|l| l[2]}
    content.should == @body1.split("\n")
  end

  it "posts multiple lines from a log file with a given seconds since epoch" do
    time = (Time.now.to_i - 5) * 1000
    response = post_with_body("/api/v1/logs/#{@log_id}/agents/#{@agent_id}?api_key=ignored&time=#{time}", @body2)
    response[:code].should == 200
    
    response = get("/api/v1/logs/#{@log_id}/agents/#{@agent_id}?api_key=ignored")
    lines =  parse_json(response[:body])
    ids = lines.map {|l| l[0]}
    times = lines.map {|l| l[1]}
    content = lines.map {|l| l[2]}

    lines_with_specified_time = @body2.split("\n")

    times.slice(0, lines_with_specified_time.size).each {|t| t.should == time}
    content.should == lines_with_specified_time + @body1.split("\n")
  end

  it "returns log lines from multiple days" do
    one_day_ago = (Time.now.to_i - (60 * 60 * 24)) * 1000
    now = Time.now.to_i * 1000
    log_id = "multiple_#{now}"
    agent_id = "10.0.2.2"
    now_lines = ["this is now", "and this is later"]
    one_day_ago_lines = ["from yesterday", "some log lines that are old", "and some more old stuff"]

    response = post_with_body("/api/v1/logs/#{log_id}/agents/#{agent_id}?api_key=ignored&time=#{one_day_ago}", one_day_ago_lines.join("\n"))
    response[:code].should == 200

    response = post_with_body("/api/v1/logs/#{log_id}/agents/#{agent_id}?api_key=ignored", now_lines.join("\n"))
    response[:code].should == 200

    response = get("/api/v1/logs/#{log_id}/agents/#{agent_id}?api_key=ignored&start=#{one_day_ago}")
    response[:code].should == 200
    lines =  parse_json(response[:body])
    ids = lines.map {|l| l[0]}
    times = lines.map {|l| l[1]}
    content = lines.map {|l| l[2]}

    content.should == one_day_ago_lines + now_lines
  end

  it "returns 100 lines from a log file (newest first line to oldest last line) by default"
  it "takes a limit parameter to return more than 100 lines"
  it "returns a max of 10,000 lines"
  it "returns lines from a log file starting from a given seconds since epoch time"
  it "returns lines from a log file between a given starting and ending since epoch time"

  it "creates a time series that has a point with the character count for each log line" do
    line1 = "this is a line to test some character length stuff"
    line2 = "how about another line?"
    body = [line1, line2].join("\n")
    log_id = "counter_test_#{Time.now.to_i}"
    agent_id = "1"

    response = post_with_body("/api/v1/logs/#{log_id}/agents/#{agent_id}?api_key=ignored", body)
    response[:code].should == 200

    time_series_id = "counters_#{log_id}_#{agent_id}"
    response = get("/api/v1/time_series/#{time_series_id}?api_key=ignored")
    response[:code].should == 200
    json = parse_json(response[:body])

    lengths = json.map {|v| v.last}
    lengths.should == [line1.length.to_f, line2.length.to_f]
  end

  it "returns a specific line id from a log with a given window of time surrounding the line"

  it "returns the lines since a given line id" do
    log_id = "test_since_line_id_#{Time.now.to_i}"
    agent_id = "127.0.0.1"

    older_lines = ["old 1", "old line 2"]
    response = post_with_body("/api/v1/logs/#{log_id}/agents/#{agent_id}?api_key=ignored", older_lines.join("\n"))
    response[:code].should == 200

    response = get("/api/v1/logs/#{log_id}/agents/#{agent_id}?api_key=ignored")
    response[:code].should == 200
    lines =  parse_json(response[:body])
    ids = lines.map {|l| l[0]}
    times = lines.map {|l| l[1]}
    content = lines.map {|l| l[2]}
    content.should == older_lines

    newer_lines = ["something newer", "and the latest"]
    response = post_with_body("/api/v1/logs/#{log_id}/agents/#{agent_id}?api_key=ignored", newer_lines.join("\n"))
    response[:code].should == 200

    last_line_id = ids.last
    last_line_time = times.last

    response = get("/api/v1/logs/#{log_id}/agents/#{agent_id}/since_time/#{last_line_time}/and_id/#{last_line_id}?api_key=ignored")
    response[:code].should == 200
    lines =  parse_json(response[:body])
    ids = lines.map {|l| l[0]}
    times = lines.map {|l| l[1]}
    content = lines.map {|l| l[2]}
    content.should == newer_lines
  end
end

describe "testing parsing rules" do
  before(:all) do
    @organization = {:id => 1}
    put_url("http://localhost:4567/api/v1/organizations", @organization.to_json)
  end

  it "will check parsing rules against a line" do
    test_rule = {
      :test_line => "here is a test 1338161212 and a date 2012-05-27 and more stuff 0.231",
      :regex => "(\\d+).*(\\d\\d\\d\\d)-(\\d\\d)-(\\d\\d).*(\\d+\\.\\d+)",
      :group_names => ["timestamp", "year", "month", "day", "response_time"]
    }
    response = post_with_body("/api/v1/parsing_rules/test?api_key=ignored", test_rule.to_json)
    response[:code].should == 200

    json = parse_json(response[:body])
    json["timestamp"].should == "1338161212"
    json["year"].should == "2012"
    json["month"].should == "05"
    json["day"].should == "27"
    json["response_time"].should == "0.231"
  end

  it "returns an error if the regex is invalid" do
    test_rule = {
      :test_line => "won't work",
      :regex => "(\\d+).*(\\d\\d\\d\\d-(\\d\\d)-(\\d\\d).*(\\d+\\.\\d+)",
      :group_names => ["timestamp", "year", "month", "day", "response_time"]
    }
    response = post_with_body("/api/v1/parsing_rules/test?api_key=ignored", test_rule.to_json)
    response[:code].should == 200

    json = parse_json(response[:body])
    json["message"].should == "Unclosed group near index 42\n(\\d+).*(\\d\\d\\d\\d-(\\d\\d)-(\\d\\d).*(\\d+\\.\\d+)\n                                          ^"
  end

  it "returns an error if the regex doesn't match against the test line" do
    test_rule = {
      :test_line => "this won't match",
      :regex => "(\\d+).*(\\d\\d\\d\\d)-(\\d\\d)-(\\d\\d).*(\\d+\\.\\d+)",
      :group_names => ["timestamp", "year", "month", "day", "response_time"]
    }
    response = post_with_body("/api/v1/parsing_rules/test?api_key=ignored", test_rule.to_json)
    response[:code].should == 200

    json = parse_json(response[:body])
    json["message"].should == "regex didn't match against provided test line"
  end
end

describe "creating time series from parsing rules" do
  before(:all) do
    @organization = {:id => 1}
    put_url("http://localhost:4567/api/v1/organizations", @organization.to_json)
  end

  def mock_parse_rules(rules, log_id, agent_id)
    put_url("http://localhost:4567/api/v1/agents/#{agent_id}/logs/#{log_id}/parse_rules", rules.to_json)
  end

  it "should create a time series using now as the timestamp and a parsed value" do
    log_id = "2_#{Time.now.to_i}"
    agent_id = 1
    time_series_name = "timestamp_now_test"
    rules = [
      {
        "regex" => "(\\d+)",
        "group_names" => ["value"],
        "outputs" => [
          {
            "time_series_name" => time_series_name,
            "key_off_agent" => true,
            "key_off_log" => true,
            "value_group_name" => "value"
          }
        ]
      }
    ]
    mock_parse_rules(rules, log_id, agent_id)

    log_lines = [
      "value one is 23 and stuff",
      "and value two is 12"
    ].join("\n")

    start_time = Time.now.to_f * 1000 / 1
    response = post_with_body("/api/v1/logs/#{log_id}/agents/#{agent_id}?api_key=ignored", log_lines)
    puts_response(response)
    response[:code].should == 200
    stop_time = Time.now.to_f * 1000 / 1

    response = get("/api/v1/time_series/#{agent_id}_#{log_id}_#{time_series_name}?api_key=ignored")
    response[:code].should == 200
    json = parse_json(response[:body])

    times = json.map {|vals| vals[0]}
    response_values = json.map {|vals| vals[1]}
    response_values.should == [23.0, 12.0]
    times.each do |time|
      time.should >= start_time
      time.should <= stop_time
    end
  end

  it "should create a time series using a parsed timestamp and a parsed value"
  it "should create a time series with a default value of 1"
  it "should create a time series without keying off agent id"
  it "should create a time series without keying off log id"
  it "should create a time series keyed off parsed group name values"
  it "should create multiple time series from a single log line"
end
