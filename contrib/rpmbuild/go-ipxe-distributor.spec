%define systemd_dir     /lib/systemd/system
%define sysconfig_dir   /etc/sysconfig

Name:           go-ipxe-distributor
Version:        1.0.0
Release:        1%{?dist}
Summary:        A webservice to provide iPXE configuration based on MAC,serial or group name - Go implementation and the successor of ipxe-distributor

Group:          System Environment/Daemons
License:        GPL
URL:            https://ypbind.de/maus/projects/go-ipxe-distributor/index.html
Source0:        https://git.ypbind.de/cgit/go-ipxe-distributor/snapshot/go-ipxe-distributor-1.0.0.tar.gz
Source1:        go-ipxe-distributor.service
Source2:        go-ipxe-distributor.default
BuildRequires:  golang

# don't build debuginfo package
%define debug_package %{nil}

# Don't barf on missing build_id
%global _missing_build_ids_terminate_build 0

%description
A webservice to provide iPXE configuration based on MAC,serial or group name - Go implementation and the successor of ipxe-distributor

%prep
%setup -q


%build
make %{?_smp_mflags}


%install
make install DESTDIR=%{buildroot}
mkdir -m 0755 -p %{buildroot}%{systemd_dir}
%{__install} -p -D -m 0644 %{SOURCE1} %{buildroot}%{systemd_dir}/%{name}.service
%{__install} -p -D -m 0644 %{SOURCE2} %{buildroot}%{sysconfig_dir}/%{name}

%post
if [ $1 == 1 ]; then
    /bin/systemctl -q enable %{name}.service
fi

%preun
if [ $1 = 0 ]; then
    /bin/systemctl stop %{name}.service >/dev/null 2>&1
    /bin/systemctl -q disable %{name}.service
fi

%postun
if [ $1 == 2 ]; then
    /bin/systemctl reload %{name}.service
fi


%files
%defattr(-,root,root,-)
%doc LICENSE README.md
%{_bindir}/ipxe_distributor
%{systemd_dir}/%{name}.service
%{sysconfig_dir}/%{name}

%changelog
* Thu Jul 02 2020 Andreas Maus <andreas.maus@atos.net> - 1.0.0
* Initial release 1.0.0

