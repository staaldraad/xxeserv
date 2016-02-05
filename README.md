XXE-FTP
----

Basic FTP server to receive payloads from instances of XXE. Keeps listening until you shut it down. Java connections shouldn't hang connecting to this either.

Usage
---
Built for Linux, so use ./xxeftp -p 2121

There are multiple modes. The server can host both FTP and HTTP, thus making it capable of serving the DTD and receiving the FTP payload.

To start the web-server (off by default) use `-w`
``` ./xxeftp -w ```
To change the web-port, use `-wp`. To Change the FTP port, use `-p`. The DTD is served out of the CWD by default. To change, use `-wd`.

To save the data received via FTP to file, use `-o filename`. The app will automatically create the file if it doesn't exist.


To build:
---
``` go build ```


