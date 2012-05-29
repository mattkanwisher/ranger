GOROOT=/var/lib/jenkins/go
GOBIN=/var/lib/jenkins/bin
PATH=$PATH:/var/lib/jenkins/bin
go build src/local_agent.go
go get launchpad.net/gocheck
go get github.com/kless/goconfig/config
cp local_agent packages/deb_pkg/errplane/usr/local/bin/errplane-local-agent
cd packages/deb_pkg
dpkg --build errplane ./
