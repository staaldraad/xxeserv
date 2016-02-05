package main

import (
    "net"
    "fmt"
    "flag"
    "os"
    "bytes"
    "io"
    "strings"
)


func parseConn(conn *net.TCPConn){
    writer := io.Writer(conn)

    writer.Write([]byte("220 Staal XXE-FTP\r\n"))

    buf := &bytes.Buffer{}
    reserved := []string{"TYPE","EPSV","EPRT"}
    for {
        data := make([]byte, 2048)
        n, err := conn.Read(data)

        if err != nil {
            fmt.Println("[x] Connection Closed")
            break
        }

        buf.Write(data[:n])

        if buf.Len() > 4 {
            cmd := string(buf.Bytes()[:4])
            if cmd == "USER" || cmd == "PASS" {
                fmt.Printf("%s: %s",cmd,string(buf.Bytes()[4:]))
                writer.Write([]byte("331 password please - version check\r\n"))
            } else if cmd == "QUIT" {
                writer.Write([]byte("221 Goodbye.\r\n"))
                break
            } else if cmd == "RETR" {
                fmt.Printf("%s\n",string(buf.Bytes()[4:]))
                writer.Write([]byte("451 Nope\r\n"))
                writer.Write([]byte("221 Goodbye.\r\n"))
                break
            } else {
                if string(buf.Bytes()[:3]) == "CWD" {
                    writer.Write([]byte("230 more data please!\r\n"))
                    fmt.Printf("%s",strings.Replace(string(buf.Bytes()[4:]),"\r\n","",1))
                } else if contains(reserved,string(buf.Bytes()[:4])) == true {
                    writer.Write([]byte("230 more data please!\r\n"))
                } else {
                    writer.Write([]byte("230 more data please!\r\n"))
                    fmt.Printf("%s\n",string(buf.Bytes()[:4]))
                }
            }
        }
        buf = &bytes.Buffer{}
    }
    defer conn.Close()
}

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func handleConnection(incomming <- chan *net.TCPConn, outgoing chan <- *net.TCPConn) {
    for conn := range incomming {
        parseConn(conn)
        outgoing <- conn
    }
}

func closeConnection(incomming <- chan *net.TCPConn){
    for conn := range incomming {
        fmt.Println("[CLOSING CONNECTION]")
        conn.Close()
    }
}


func main(){
    portPtr := flag.Int("p",2121,"Port to listen on")
    flag.Parse()

    waiting, complete := make(chan *net.TCPConn), make(chan *net.TCPConn)
    var err error

    for i := 0; i < 1 ; i++ {
        go handleConnection(waiting, complete)
    }
    go closeConnection(complete)

    var clientConn *net.TCPConn
    addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprint(":",*portPtr))
    ls, err := net.ListenTCP("tcp",addr)
    if err != nil {
        fmt.Println("[x] - Failed to start connection\n",err)
        os.Exit(1)
    }
    fmt.Println("GO XXE FTP Server - Port: ",*portPtr)

    for {
        clientConn, err = ls.AcceptTCP()
        if err != nil {
            fmt.Println("[x] - Failed to accept connection\n",err)
            os.Exit(1)
        }
        fmt.Printf("[*] Connection Accepted from [%s]\n",clientConn.RemoteAddr().String())
        waiting <- clientConn
    }
}
