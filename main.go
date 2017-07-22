package main

import (
    "bufio"
    "bytes"
    "crypto/tls"
    "crypto/x509"
    "encoding/binary"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os/exec"
    "strings"
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

func recv(conn *tls.Conn) []byte {
    // log.Println("Entered recv()")
    x := make([]byte, 4)
    // log.Println("About to READ")
    conn.SetDeadline(time.Now().Add(5 * time.Second))
    _, err := conn.Read(x)
    if err != nil {
        panic(err)
    }
    var parsedLength int32
    buf := bytes.NewBuffer(x)
    readerr := binary.Read(buf, binary.BigEndian, &parsedLength)
    if readerr != nil {
        panic(readerr)
    }
    // log.Printf("wtf is anything %d", parsedLength)
    data := make([]byte, parsedLength)
    fuckit, err := conn.Read(data)
    if err != nil {
        panic(err)
    }
    // log.Println(fuckit)
    // log.Printf("parsed msg: %s", data[:fuckit])

    return data[:fuckit]
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
    // log.Println(msgFinal)

    conn.Write(bytes.NewBufferString(msgFinal).Bytes())
}

func pull(conn *tls.Conn) {
    msg := mkMessage("Public", "ae0a6853-2b68-469d-a81c-fc5e5ab3afb5", "jackjrabbit")
    msg["type"] = "sync"
    msgFinal := finalizeMessage(msg)
    conn.SetDeadline(time.Now().Add(5 * time.Second))
    conn.Write(bytes.NewBufferString(msgFinal).Bytes())
}

type taskResponse struct {
    SyncKey string
    Tasks   []task
    Status  string
    Code    int
}

func parseResponse(resp []byte) taskResponse {
    buff := bytes.NewBuffer(resp)
    scanner := bufio.NewScanner(buff)
    var tasks []task
    var headers [][]string
    for scanner.Scan() {
        text := scanner.Text()
        var mytask task
        if len(text) < 1 {
            continue
        } else if strings.Split(text, "")[0] == "{" {
            log.Println("smells like JSON")
            log.Println(text)
            json.Unmarshal(scanner.Bytes(), &mytask)
            tasks = append(tasks, mytask)
        } else if strings.Contains(text, ":") {
            xyz := strings.Split(text, ":")
            headers = append(headers, xyz)
        } else if len(text) == 36 {
            log.Println("found synckey uuid, maybe?")
            log.Println(text)
        }
        // log.Println()
        // log.Println(len(tasks))
    }
    // log.Printf("parseResponse: %s")
    log.Println(headers)
    return taskResponse{}
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
    // log.Printf("FinalizeMessage: Got %v length", length)
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
    // stats(conn)
    pull(conn)
    resp := recv(conn)
    parseResponse(resp)

    conn.Close()
}
