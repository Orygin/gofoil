package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var rootPath = flag.String("root","Z:\\", "Path to start scanning from")
var scanNodesArg = flag.String("folders", "Downloads,Games/switch", "CSV of folders to scan")

var scanNodes []string

var hostIP = flag.String("ip", "192.168.1.95", "Host IP to display to switch")
var hostPort = flag.String("port", "8000", "Host Port to display to switch")

const htmlpage =`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>Select switch IP to connect to</title><style>p, #switch {font-size: 8vw;}button{width:100%;height: 80px;background-color: #EFA6A2;font-size: 3vw;}</style></head>
<body><p>select switch IP to connect to</p><form action="#" method="post"><input type="text" id="switch" name="switch" value="192.168.1.17"><button type="submit">go</button></form></body></html>`

var filesMap map[string]string

func main(){
	initFlags()
	filesMap = make(map[string]string)
    r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)

	r.PathPrefix("/files/").Handler(http.StripPrefix("/files/", http.FileServer(http.Dir(*rootPath))))
    r.Use(loggingMiddleware)

    log.Printf("Starting server at %s:%s, with root %s + %s", *hostIP, *hostPort, *rootPath, *scanNodesArg)

	srv := &http.Server{
		Handler:      r,
		Addr:         *hostIP+":"+*hostPort,
		// Good practice: enforce timeouts for servers you create!
		// But here the switch will be downloading file, slowly...
		//WriteTimeout: 15 * time.Second,
		//ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
func initFlags () {
	flag.Parse()
	scanNodes = strings.Split(*scanNodesArg, ",")
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		// Display web page to choose switch ip
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlpage))
		return
	} else if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(500)
		return
	}
	destination := r.Form.Get("switch")
	if destination == "" {
		w.WriteHeader(400)
	}
	log.Println(destination)

	files := []string{}
	length := 0
	// Find NSP files within the scanNodes list of directories
	for _, node := range scanNodes {
		err := filepath.Walk(*rootPath + node, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			switch filepath.Ext(info.Name()) {
			case ".nsp":
				fallthrough
			case ".nsz":
				fallthrough
			case ".xci":
				//relPath := filepath.Base(path)
				relPath := strings.TrimPrefix(path, *rootPath)
				relPath = strings.Replace(relPath, "\\", "/", -1) // Remove ugly windows \ seperator
				finalPath := fmt.Sprintf("%s:%s/files/%s\n", *hostIP, *hostPort, url.PathEscape(relPath))
				files = append(files, finalPath)
				length += len(finalPath)
				filesMap[relPath] = path
			}
			return nil
		})
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}
	}

	// create socket connection to switch
	conn, err := net.Dial("tcp", destination + ":2000")
	if err != nil {
		log.Printf("%v", err)
		return
	}

	defer conn.Close()

	sendNSPList(files, conn, length)

	buff := make([]byte, 1024)
	for n, err := conn.Read(buff); n< 0 && err != nil; {
		log.Println("rcv")
		time.Sleep(time.Millisecond*10)
	}


	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlpage))
}

// Taken from : https://github.com/bycEEE/tinfoilusbgo/blob/master/main.go. Adapted to work for network (bigEndian)
// sendNSPList creates a payload out of an NSPList struct and sends it to the switch.
func sendNSPList(fileList []string, out io.Writer, length int) {

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(length)) // NSP list length
	out.Write(buf)


	fmt.Printf("Sending NSP list: %v\n", fileList)

	for _, path := range fileList {
		buf = make([]byte, len(path))
		copy(buf, path) // File path followed by newline
		out.Write(buf)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Printf("%s : %s", r.RequestURI, r.Header.Get("Range"))
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}