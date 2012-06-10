CONTROL=packages/deb_pkg/errplane/DEBIAN/control 
VER_SRC=src/local_agent.go
GOROOT=/var/lib/jenkins/go
GOBIN=/var/lib/jenkins/bin
PATH=$PATH:/var/lib/jenkins/bin
rm ./packages/deb_pkg/errplane*.deb
rm ./packages/deb_pkg/errplane-local-agent*
#requires gnu sed 
sed -i "s/_BUILD_/${BUILD_NUMBER}/g" $CONTROL 
sed -i "s/_BUILD_/${BUILD_NUMBER}/g" $VER_SRC 
go get launchpad.net/gocheck
go get github.com/kless/goconfig/config
go get github.com/droundy/goopt
go build src/local_agent.go
#cp local_agent packages/deb_pkg/errplane/usr/local/bin/errplane-local-agent
rm packages/deb_pkg/errplane/usr/local/bin/errplane-local-agent
cp local_agent packages/deb_pkg/errplane-local-agent-${BUILD_NUMBER}
cd packages/deb_pkg
dpkg --build errplane ./
cd ..

