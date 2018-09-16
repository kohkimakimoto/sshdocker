Name:           %{_product_name}
Version:        %{_product_version}

Release:        1.el%{_rhel_version}
Summary:        sshdocker
Group:          Development/Tools
License:        MIT
Source0:        %{name}_linux_amd64.zip
Source1:        sshdocker.yml
Source2:        sshdocker.sysconfig
Source3:        sshdocker.service
BuildRoot:      %(mktemp -ud %{_tmppath}/%{name}-%{version}-%{release}-XXXXXX)

%description
Run docker containers over SSH.

%prep
%setup -q -c

%install
mkdir -p %{buildroot}/%{_bindir}
cp %{name} %{buildroot}/%{_bindir}

mkdir -p %{buildroot}/%{_sysconfdir}/%{name}
cp %{SOURCE1} %{buildroot}/%{_sysconfdir}/sshdocker.yml

mkdir -p %{buildroot}/%{_sysconfdir}/sysconfig
cp %{SOURCE2} %{buildroot}/%{_sysconfdir}/sysconfig/sshdocker

mkdir -p %{buildroot}/var/lib/sshdocker

%if 0%{?fedora} >= 14 || 0%{?rhel} >= 7
mkdir -p %{buildroot}/%{_unitdir}
cp %{SOURCE3} %{buildroot}/%{_unitdir}/
%endif

%if 0%{?fedora} >= 14 || 0%{?rhel} >= 7
%post
%systemd_post sshdocker.service

%preun
%systemd_preun sshdocker.service
%endif

%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root,-)
%attr(755, root, root) %{_bindir}/%{name}
%config(noreplace) %{_sysconfdir}/sshdocker.yml
%config(noreplace) %{_sysconfdir}/sysconfig/sshdocker

%if 0%{?fedora} >= 14 || 0%{?rhel} >= 7
%{_unitdir}/sshdocker.service
%endif

%doc
