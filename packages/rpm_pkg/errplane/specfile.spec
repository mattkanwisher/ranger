Summary: Errplane Local Agent
Name: errplane
Version: _BUILD_
Release: 1
License: Commercial
Group: Applications/System


%description
Brief description of software package.

%prep

%build

%install
rm /usr/local/errplane/errplane-local-agent; true
ln -s /usr/local/errplane/errplane-local-agent-_BUILD_ /usr/local/errplane/errplane-local-agent

%clean

%files
%defattr(-,root,root)

%config /etc/errplane.conf
/var/log/errplane/errplane.log
/var/run/errplane/errplane.pid
/usr/local/errplane/errplane-local-agent-_BUILD_
%exclude /usr/local/errplane/.gitkeep