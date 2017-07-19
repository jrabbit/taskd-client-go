package main

import (
    "crypto/tls"
    "crypto/x509"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os/exec"
    "time"
    "unsafe"
)

type task struct {
    // https://github.com/manishrjain/taskreview/blob/master/task.go#L17
    Completed   string   `json:"end,omitempty"`
    Created     string   `json:"entry,omitempty"`
    Description string   `json:"description,omitempty"`
    Modified    string   `json:"modified,omitempty"`
    Project     string   `json:"project,omitempty"`
    Status      string   `json:"status,omitempty"`
    Tags        []string `json:"tags,omitempty"`
    Uuid        string   `json:"uuid,omitempty"`
    // Xid         string   `json:"xid,omitempty"`
    Reviewed string  `json:"reviewed,omitempty"`
    Urgency  float64 `json:"urgency,omitempty"`
}

func version() string {
    return "v0.0.1a"
}

func json_read() {
    cmd := exec.Command("task", "export")
    out, err := cmd.Output()
    if err != nil {
        panic(err)
    }
    var tasks []task
    json.Unmarshal(out, &tasks)
    fmt.Println(tasks[0].Description)
}

func recv(conn *tls.Conn) {
    log.Println("Entered recv()")
    x := make([]byte, 4)
    log.Println("About to READ")
    conn.SetDeadline(time.Now().Add(5 * time.Second))
    length, err := conn.Read(x)
    if err != nil {
        panic(err)
    }
    log.Println(length)

}

func mkMessage(org string, uuid string, user string) map[string]string {
    var headers = map[string]string{}
    headers["client"] = fmt.Sprintf("taskc-go %s", version())
    headers["protocol"] = "v1"
    return headers
}

func stats() {

}

func finalizeMessage(msg string) string {
    length := unsafe.Sizeof(msg)
    length += 4
    return string(length) + msg
}

func main() {
    // First, create the set of root certificates. For this example we only
    // have one. It's also possible to omit this in order to use the
    // default root set of the current operating system.
    log.Println("Entered main()")

    roots := x509.NewCertPool()
    cacert, err := ioutil.ReadFile("/home/jack/.task/beta.getpizza.cat.ca.cert.pem")
    if err != nil {
        panic(err)
    }
    ok := roots.AppendCertsFromPEM(cacert)
    if !ok {
        panic("failed to parse root certificate")
    }

    cert, err := tls.LoadX509KeyPair("/home/jack/.task/pizzacat-jackjrabbit.cert.pem", "/home/jack/.task/pizzacat-jackjrabbit.key.pem")
    if err != nil {
        panic(err)
    }
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      roots,
    }
    tlsConfig.BuildNameToCertificate()

    conn, err := tls.Dial("tcp", "beta.getpizza.cat:53589", tlsConfig)
    if err != nil {
        panic("failed to connect: " + err.Error())
    }
    x := mkMessage("Public", "be3e0803-cb00-4803-b103-1493b89a1302", "jack")
    log.Println(x)
    // conn.Write()
    // log.Println(finalizeMessage(x))
    recv(conn)

    conn.Close()
}
