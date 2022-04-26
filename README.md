# OpenConnect Auth Hack
Changes Default U/P auth for Openconnect to use Username &amp; Password on one screen.

EX: VPN CLIENT ----> OC Auth Hack (443) ----> OCSERV (4433)

1. Install GO & ensure Openconnect TCP & UDP port is set to a custom port (EX: 4433)
2. Port forward TCP/UDP 443 4433 & ensure firewall is configured to port forward these
3. Around line 55, change CERT and KEY to match the certificate & key values located in ocserv

```cert, _ := tls.LoadX509KeyPair("CERT.PEM", "KEY.PEM")```

4. Around line 78 Point The OC Auth Hack to the OC SERV, **ensure your using the domain that the cert is issued for**

```remoteConn, err := tls.Dial("tcp", "MYSERVER.COM:4433", nil) //remote can be unix cleartext socket of ocserv```

5. Test by running ```go run main.go``` If you see no output you are good, try and connect to ```MYSERVER.COM:443``` (not 4433 OCSERV)

** Linux Service **

1. ```mkdir /etc/oc_hack```
2. Copy the service.sh & main.go to /etc/oc_hack then chmod +x service.sh
