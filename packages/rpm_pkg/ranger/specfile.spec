Summary: Errplane Local Agent
Name: ranger
Version: _BUILD_
Release: 1
License: Commercial
Group: Applications/System


%description
Brief description of software package.

%prep

%build

%install
rm /usr/local/ranger/ranger-local-agent; true
ln -s /usr/local/ranger/ranger-local-agent-_BUILD_ /usr/local/ranger/ranger-local-agent

%clean

%files
%defattr(-,root,root)

%config /etc/ranger.conf
/etc/init.d/ranger
/var/log/ranger/ranger.log
/var/run/ranger/ranger.pid
/usr/local/ranger/ranger-local-agent-_BUILD_
%exclude /usr/local/ranger/.gitkeep