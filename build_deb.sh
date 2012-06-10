#!/bin/bash
CONTROL=packages/deb_pkg/errplane/DEBIAN/control 
CONTROL_POST=packages/deb_pkg/errplane/DEBIAN/postinst 
VER_SRC=src/local_agent.go
GOROOT=/var/lib/jenkins/go
GOBIN=/var/lib/jenkins/bin
PATH=$PATH:/var/lib/jenkins/bin

function do_build() {
  BUILD_CPU= $1
  NEW_BUILD_NUMBER=1.0.${BUILD_NUMBER}-${BUILD_CPU}
  rm ./packages/deb_pkg/errplane*.deb
  rm ./packages/deb_pkg/errplane-local-agent*
  #requires gnu sed 
  sed -i "s/_OS_/${BUILD_CPU}/g" $CONTROL 
  sed -i "s/_BUILD_/${NEW_BUILD_NUMBER}/g" $CONTROL 
  sed -i "s/_BUILD_/${NEW_BUILD_NUMBER}/g" $CONTROL_POST
  sed -i "s/_BUILD_/${NEW_BUILD_NUMBER}/g" $VER_SRC 
  go get launchpad.net/gocheck
  go get github.com/kless/goconfig/config
  go get github.com/droundy/goopt
  go build src/local_agent.go
  #cp local_agent packages/deb_pkg/errplane/usr/local/bin/errplane-local-agent
  chmod +x local_agent
  rm packages/deb_pkg/errplane/usr/local/errplane/errplane-local-agent*
  GOARCH=${BUILD_CPU} GOOS=linux cp local_agent packages/deb_pkg/errplane/usr/local/errplane/errplane-local-agent-${NEW_BUILD_NUMBER}
  cd packages/deb_pkg
  dpkg --build errplane ./
  cd ../..
  sha=`shasum -a 256 local_agent`
  echo "SHA 256 - ${sha}"
}

do_build amd64

do_build i386