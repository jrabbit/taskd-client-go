package main

import (
    "bytes"
    "crypto/tls"
    "crypto/x509"
    "encoding/binary"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os/exec"
    "text/template"
    "time"
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
    _, err := conn.Read(x)
    if err != nil {
        panic(err)
    }
    var parsedLength int32
    buf := bytes.NewReader(x)
    readerr := binary.Read(buf, binary.BigEndian, &parsedLength)
    if readerr != nil {
        panic(readerr)
    }
    // log.Println(buf)
    // newbuf := new(bytes.Buffer)
    data := make([]byte, 1280)
    fuckit, err := conn.Read(data)
    if err != nil {
        panic(err)
    }
    log.Println(fuckit)
    xyz := bytes.NewBuffer(data[fuckit:])
    log.Println(xyz.String())
    // log.Println(newbuf.String())
    // log.Printf("demonic screaming %s", newbuf.String())

}

func mkMessage(org string, uuid string, user string) map[string]string {
    var headers = map[string]string{}
    headers["client"] = fmt.Sprintf("taskc-go %s", version())
    headers["protocol"] = "v1"
    headers["org"] = org
    headers["key"] = uuid
    headers["user"] = user
    return headers
}

func stats(conn *tls.Conn) {
    msg := mkMessage("Public", "be3e0803-cb00-4803-b103-1493b89a1302", "jack")
    msg["type"] = "statistics"
    msgFinal := finalizeMessage(msg)
    conn.SetDeadline(time.Now().Add(5 * time.Second))
    log.Println(msgFinal)

    conn.Write(bytes.NewBufferString(msgFinal).Bytes())
}

func finalizeMessage(msg map[string]string) string {
    tmpl, err := template.New("test").Parse("client: {{.client}}\ntype: {{.type}}\nprotocol: {{.protocol}}\nuser: {{.user}}\norg: {{.org}}\nkey: {{.key}}\n\n")
    if err != nil {
        panic(err)
    }
    buf := new(bytes.Buffer)

    err = tmpl.Execute(buf, msg)
    if err != nil {
        panic(err)
    }
    x := buf.String()
    fmt.Println(x)
    length := len(x)
    log.Printf("FinalizeMessage: Got %v length", length)
    length += 4

    buf2 := new(bytes.Buffer)
    len32 := int32(length)

    binary.Write(buf2, binary.BigEndian, len32)
    log.Println(buf2.String())
    return buf2.String() + x
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

    // log.Println(x)
    // conn.Write()
    // log.Println(finalizeMessage(x))
    stats(conn)
    recv(conn)

    conn.Close()
}
