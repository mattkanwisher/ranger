/var/lib/jenkins/bin/go build src/local_agent.go
go install launchpad.net/gocheck
go install 
cp local_agent packages/deb_pkg/errplane/usr/local/bin/errplane-local-agent
cd packages/deb_pkg
dpkg --build errplane ./
