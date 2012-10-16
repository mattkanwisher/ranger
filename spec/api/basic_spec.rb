require "#{File.dirname(__FILE__)}/../spec_helper.rb"

describe "basic testing of local agent" do
  before(:all) do
  end

  it "spawns a new instance and writes a pid file" do
    s = IO.read("/tmp/ranger.pid")
    s.should_not be_empty
  end
end