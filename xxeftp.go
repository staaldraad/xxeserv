package main

import (
    "net"
    "fmt"
    "flag"
    "bytes"
    "io"
    "strings"
    "net/http"
    "log"
    "os"
)

var logger *log.Logger
var fileLogger *log.Logger
var hostDir string = "./"

func parseConn(conn *net.TCPConn){
    writer := io.Writer(conn)
    writer.Write([]byte("220 Staal XXE-FTP\r\n"))
    var olog *log.Logger

    if fileLogger != nil {
        olog = fileLogger
    } else {
        olog = log.New(os.Stderr, "", 0)
    }

    buf := &bytes.Buffer{}
    reserved := []string{"TYPE","EPSV","EPRT"}
    for {
        data := make([]byte, 2048)
        n, err := conn.Read(data)

        if err != nil {
            logger.Println("[x] Connection Closed")
            break
        }

        buf.Write(data[:n])

        if buf.Len() > 4 {
            cmd := string(buf.Bytes()[:4])
            if cmd == "USER" || cmd == "PASS" {
                olog.Printf("%s: %s",cmd,string(buf.Bytes()[4:]))
                writer.Write([]byte("331 password please - version check\r\n"))
            } else if cmd == "QUIT" {
                writer.Write([]byte("221 Goodbye.\r\n"))
                break
            } else if cmd == "RETR" {
                olog.Printf("%s",string(buf.Bytes()[4:]))
                writer.Write([]byte("451 Nope\r\n"))
                writer.Write([]byte("221 Goodbye.\r\n"))
                break
            } else {
                if string(buf.Bytes()[:3]) == "CWD" {
                    writer.Write([]byte("230 more data please!\r\n"))
                    olog.Printf("/%s",strings.Replace(string(buf.Bytes()[4:]),"\r\n","",1))
                } else if contains(reserved,string(buf.Bytes()[:4])) == true {
                    writer.Write([]byte("230 more data please!\r\n"))
                } else {
                    writer.Write([]byte("230 more data please!\r\n"))
                    olog.Printf("%s\n",string(buf.Bytes()[:4]))
                }
            }
        }
        buf = &bytes.Buffer{}
    }
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
        logger.Println("[*] Closing FTP Connection")
        conn.Close()
    }
}

func logRequest(w http.ResponseWriter, req *http.Request){
    if _, err := os.Stat(fmt.Sprintf("%s/%s",hostDir,req.URL.Path)); err != nil {
        logger.Printf("[%s][404] %s\n",req.RemoteAddr,req.URL)
        fmt.Fprintf(w,"Not Found")
    } else {
        logger.Printf("[%s][200] %s\n",req.RemoteAddr,req.URL)
        if req.URL.Path == "/" {
            http.ServeFile(w,req,fmt.Sprintf("%s/",hostDir))
        } else if req.URL.Path[len(req.URL.Path)-1:] == "/" {
            http.ServeFile(w,req,fmt.Sprintf("%s/%s",hostDir,req.URL.Path[:len(req.URL.Path)-1]))
        } else {
            http.ServeFile(w,req,fmt.Sprintf("%s/%s",hostDir,req.URL.Path))
        }
    }
}


func serveWeb(port int,dir string){
    logger.Printf("[*] Starting Web Server on %d [%s]\n",port,dir)
    hostDir = dir
    http.HandleFunc("/",logRequest)
    go http.ListenAndServe(fmt.Sprint(":",port), nil)
}


func main(){
    portPtr := flag.Int("p",2121,"Port to listen on")
    webEnabledPtr := flag.Bool("w",false,"Setup web-server for DTDs")
    webPortPtr := flag.Int("wp",2122,"Port to serve DTD on")
    webFolderPtr := flag.String("wd","./","Folder to server DTD(s) from")
    fileLog := flag.String("o","","File location to log to")
    flag.Parse()

    logger = log.New(os.Stderr, "", log.LstdFlags)

    if *fileLog != "" {
        if _, err := os.Stat(*fileLog); os.IsNotExist(err) {
            logger.Println("[*] File doesn't exist, creating")
            if _, err:= os.Create(*fileLog); err != nil {
                logger.Fatal("[x] Unable to create log file! Exiting.")
            }
        }
        errorlog, err := os.OpenFile(*fileLog,  os.O_RDWR, 0666)
        if err != nil {
            logger.Fatal("error opening file: %v", err)
        }
        fileLogger = log.New(errorlog, "", 0)
        logger.Printf("[*] Storing session into the file: %s",*fileLog)
    }


    if *webEnabledPtr == true {
        serveWeb(*webPortPtr,*webFolderPtr)
    }

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
        logger.Fatal("[x] - Failed to start connection\n",err)
    }

    logger.Println("[*] GO XXE FTP Server - Port: ",*portPtr)

    for {
        clientConn, err = ls.AcceptTCP()
        if err != nil {
            logger.Fatal("[x] - Failed to accept connection\n",err)
        }
        logger.Printf("[*] Connection Accepted from [%s]\n",clientConn.RemoteAddr().String())
        waiting <- clientConn
    }
}
