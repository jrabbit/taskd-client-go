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
    "os"
    "os/exec"
    "os/user"
    "path/filepath"
    "strconv"
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
    return "v0.0.1b"
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

func readRC() map[string]string {
    usr, _ := user.Current()
    dir := usr.HomeDir
    path := filepath.Join(dir, ".taskrc")
    rc, err := os.Open(path)
    if err != nil {
        panic(err)
    }
    settings := make(map[string]string)
    scanner := bufio.NewScanner(rc)
    for scanner.Scan() {
        text := scanner.Text()
        if len(text) < 1 {
            continue
        } else if strings.Split(text, "")[0] == "#" {
            continue
        } else {
            x := strings.Split(text, "=")
            value := strings.Replace(x[1], "\\/", "/", -1)
            settings[x[0]] = value
        }
    }
    // log.Println(settings)
    return settings
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

func mkMessage(creds string) map[string]string {
    // log.Println(creds)
    splitted := strings.Split(creds, "/")
    log.Println(splitted)
    org := splitted[0]
    user := splitted[1]
    uuid := splitted[2]
    var headers = map[string]string{}
    headers["client"] = fmt.Sprintf("taskc-go %s", version())
    headers["protocol"] = "v1"
    headers["org"] = org
    headers["key"] = uuid
    headers["user"] = user
    return headers
}

func stats(conn *tls.Conn, credentials string) {
    msg := mkMessage(credentials)
    msg["type"] = "statistics"
    msgFinal := finalizeMessage(msg)
    conn.SetDeadline(time.Now().Add(5 * time.Second))
    // log.Println(msgFinal)

    conn.Write(bytes.NewBufferString(msgFinal).Bytes())
}

func pull(conn *tls.Conn, credentials string) {
    msg := mkMessage(credentials)
    msg["type"] = "sync"
    msgFinal := finalizeMessage(msg)
    conn.SetDeadline(time.Now().Add(5 * time.Second))
    conn.Write(bytes.NewBufferString(msgFinal).Bytes())
}

type taskResponse struct {
    SyncKey       string
    Tasks         []task
    Status        string
    Code          int
    ServerVersion string
    RawHeaders    map[string]string
}

func parseResponse(resp []byte) taskResponse {
    buff := bytes.NewBuffer(resp)
    scanner := bufio.NewScanner(buff)
    var tasks []task
    // var headers [][]string
    headers := make(map[string]string)
    var synckey string
    for scanner.Scan() {
        text := scanner.Text()
        var mytask task
        if len(text) < 1 {
            continue
        } else if strings.Split(text, "")[0] == "{" {
            json.Unmarshal(scanner.Bytes(), &mytask)
            tasks = append(tasks, mytask)
        } else if strings.Contains(text, ":") {
            xyz := strings.Split(text, ":")
            headers[xyz[0]] = strings.TrimLeft(xyz[1], " ")
        } else if len(text) == 36 {
            synckey = text
        }
    }
    code, err := strconv.Atoi(headers["code"])
    if err != nil {
        panic("Couldn't convert code to int " + err.Error())
    }
    parsed := taskResponse{
        SyncKey:    synckey,
        Code:       code,
        Tasks:      tasks,
        Status:     headers["status"],
        RawHeaders: headers,
    }
    log.Println(parsed)
    return parsed
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
    length := len(x)
    length += 4

    buf2 := new(bytes.Buffer)
    len32 := int32(length)

    binary.Write(buf2, binary.BigEndian, len32)
    log.Println(buf2.String())
    return buf2.String() + x
}

func connect(settings map[string]string) *tls.Conn {
    roots := x509.NewCertPool()
    cacert, err := ioutil.ReadFile(settings["taskd.ca"])
    if err != nil {
        panic(err)
    }
    ok := roots.AppendCertsFromPEM(cacert)
    if !ok {
        panic("failed to parse root certificate")
    }

    cert, err := tls.LoadX509KeyPair(settings["taskd.certificate"], settings["taskd.key"])
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
    return conn
}

func main() {
    log.Println("Entered main()")
    // // stats(conn)
    // pull(conn)
    // resp := recv(conn)
    // parseResponse(resp)
    // readRC()
    rc := readRC()
    conn := connect(rc)
    // log.Println(rc)
    stats(conn, rc["taskd.credentials"])
    resp := recv(conn)
    log.Println(parseResponse(resp))

    conn.Close()
}
