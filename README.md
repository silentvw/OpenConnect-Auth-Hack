# OpenConnect Auth Hack
Changes Default U/P auth for Openconnect to use Username &amp; Password on one screen. Tested on CentOS 7

**If you are running Certbot this service will need to be restarted along with ocserv when the cert is renewed!**

![Screenshot](https://github.com/thomaswilbur/OpenConnect-Auth-Hack/blob/main/Screen%20Shot%202022-04-26%20at%205.09.56%20AM.png?raw=true)

VPN CLIENT ----> OC Auth Hack (443) ----> OCSERV (4433 TCP)
VPN CLIENT <----------------------------- OCSERV(443 UDP)

### Installation

1. Install GO & ensure Openconnect TCP & UDP port is set to a custom port (EX: TCP: 4433 UDP 443)
2. Port forward TCP/UDP 443 & ensure firewall is configured to port forward these
3. Around line 55, change CERT and KEY to match the certificate & key values located in ocserv

```cert, _ := tls.LoadX509KeyPair("CERT.PEM", "KEY.PEM")```

4. Around line 78 Point The OC Auth Hack to the OC SERV

```remoteConn, err := tls.Dial("tcp", "127.0.0.1:4433", nil)```

5. Test by running ```go run main.go``` If you see no output you are good, try and connect to ```MYSERVER.COM:443``` (not 4433 OCSERV)

### Linux Service

1. ```mkdir /etc/oc_hack```
2. Copy the service.sh & main.go to /etc/oc_hack then chmod +x service.sh
3. Install oc_hack.service to your system
4. systemctl enable oc_hack
5. systemctl start oc_hack

### Certbot

Update your cron/service to this:

```certbot renew --quiet && systemctl restart ocserv && systemctl restart oc_hack```

