# OpenConnect Auth Hack
Changes Default U/P auth for Openconnect to use Username &amp; Password on one screen.

EX: VPN CLIENT ----> OC Auth Hack (443) ----> OCSERV (4433)

1. Install GO & ensure Openconnect TCP & UDP port is set to a custom port (EX: 4433)
2. Port forward TCP/UDP 443 4433 & ensure firewall is configured to port forward these
3. Around line 55, change CERT and KEY to match the certificate & key values located in ocserv

```cert, _ := tls.LoadX509KeyPair("CERT.PEM", "KEY.PEM")```

