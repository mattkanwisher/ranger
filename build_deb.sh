#!/bin/bash
CONTROL=packages/deb_pkg/errplane/DEBIAN/control 
CONTROL_POST=packages/deb_pkg/errplane/DEBIAN/postinst 
RPM_SPECFILE=packages/rpm_pkg/errplane/specfile.spec
VER_SRC=src/errplane/local_agent.go
JENKINS_HOME=/var/lib/jenkins
GOROOT=/var/lib/jenkins/go
GOBIN=/var/lib/jenkins/bin
PATH=$PATH:/var/lib/jenkins/bin
RPM_BUILD_ROOT=${JENKINS_HOME}/jobs/local_agent/workspace/rpm_pkg/errplane/BUILDROOT
DEB_PKG_ROOT=packages/deb_pkg/errplane

#copy out latest rpm macro file
cp packages/rpm_pkg/errplane/dot_rpm_macros ${JENKINS_HOME}/.rpmmacros

rm -rf output  ; true
mkdir output
function do_build() {
  BUILD_CPU=$1
  NEW_BUILD_NUMBER=1.0.${BUILD_NUMBER}
  #-${BUILD_CPU}


  rm ./packages/deb_pkg/errplane*.deb
  rm ./packages/deb_pkg/errplane-local-agent*
  #requires gnu sed 
  sed -i "s/_OS_/${BUILD_CPU}/g" $CONTROL 
  sed -i "s/_BUILD_/${NEW_BUILD_NUMBER}/g" $CONTROL 
  sed -i "s/_BUILD_/${NEW_BUILD_NUMBER}/g" $CONTROL_POST
  sed -i "s/_BUILD_/${NEW_BUILD_NUMBER}/g" $VER_SRC 
  sed -i "s/_BUILD_/${NEW_BUILD_NUMBER}/g" $RPM_SPECFILE
  go get launchpad.net/gocheck
  go get github.com/kless/goconfig/config
  go get github.com/droundy/goopt
  GOPATH=`pwd`:$GOPATH GOARCH=$2 GOOS=linux go build -o local_agent -v  main
  #cp local_agent packages/deb_pkg/errplane/usr/local/bin/errplane-local-agent
  chmod +x local_agent
  rm ${DEB_PKG_ROOT}/usr/local/errplane/errplane-local-agent*
  cp local_agent packages/deb_pkg/errplane/usr/local/errplane/errplane-local-agent-${NEW_BUILD_NUMBER}
  cd packages/deb_pkg
  dpkg --build errplane ./
  cd ../..
  sha=`shasum -a 256 local_agent`
  echo "SHA 256 - ${sha}"


  #TODO: sha hash of debian and rpm packages

  #TODO get i386 ver also
  rm -rf ${RPM_BUILD_ROOT}/*
  RPM_BUILD_ROOT_WITH_OS=${RPM_BUILD_ROOT}/errplane-${NEW_BUILD_NUMBER}.x86_64
  echo "Trying to create ${RPM_BUILD_ROOT_WITH_OS}"
  mkdir -p $RPM_BUILD_ROOT_WITH_OS
  echo "Trying to create ${RPM_BUILD_ROOT_WITH_OS} to DEB_PKG_ROOT"
  cp -r ${DEB_PKG_ROOT}/etc ${RPM_BUILD_ROOT_WITH_OS}/etc 
  cp -r ${DEB_PKG_ROOT}/usr ${RPM_BUILD_ROOT_WITH_OS}/usr 
  cp -r ${DEB_PKG_ROOT}/var ${RPM_BUILD_ROOT_WITH_OS}/var


  cd packages/rpm_pkg/errplane
  rpmbuild --bb specfile.spec
  cd ../../../

}

#do_build amd64 amd64

do_build i386 386

git reset HEAD