package ranger_tests

import (
  "ranger"
  "testing"
  "launchpad.net/gocheck"
)

//http://labix.org/gocheck

func Test(t *testing.T) { TestingT(t)}

type LocalAgentSuite struct{}
var _ = Suite(&LocalAgentSuite{})

func (s *LocalAgentSuite) TestShouldRequestFoo(c *C) {
  data, err ranger.dummy_method

  c.Check(err, IsNil)
  c.Assert(string(data), Equals, "This is Foo handler!")
}
