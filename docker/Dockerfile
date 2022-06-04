FROM centos:centos7.6.1810

MAINTAINER clownfish osgochina@gmail.com

### 安装公共脚本
RUN yum -y update && yum -y  install deltarpm  && \
    yum -y install epel-release && \
    yum -y update && \
    yum -y install vim  wget which net-tools bzip2 \
    && yum clean all && rm -rf /tmp/*

### 安装xfce桌面
RUN yum --enablerepo=epel -y -x gnome-keyring --skip-broken groups install "Xfce"  && \
    yum -y groups install "Fonts" && \
    yum erase -y *power* *screensaver* && \
    yum clean all && \
    rm /etc/xdg/autostart/xfce-polkit* && \
    /bin/dbus-uuidgen > /etc/machine-id

# VNC Servers
RUN  yum -y install \
       tigervnc-server \
       tigervnc-server-module \
       xinetd \
       xclip \
    && yum clean all && rm -rf /tmp/*



#vnc代理使用的端口和实训机管理端口
ENV PROXY_VNC_PORT=5899 \
    PROXY_MANAGER_PORT=5898 \
    VNC_SESSION_NAME=xfce \
    TZ=Asia/Shanghai \
    USERID=1000 \
    USER=vprix


RUN /usr/sbin/useradd $USER  -U -M -u $USERID

ENTRYPOINT ["/bin/vprix","--debug"]