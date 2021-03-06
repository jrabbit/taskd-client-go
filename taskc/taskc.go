package taskc

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

func Version() string {
	return "v0.1.1"
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

func ReadRC() (map[string]string, error) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	path := filepath.Join(dir, ".taskrc")
	rc, err := os.Open(path)
	if err != nil {
		return nil, err
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
	return settings, nil
}

func CheckCreds(settings TaskSettings) {
	if settings.Creds == "" {
		fmt.Println("No credentials specified. Cowardly exiting.")
		os.Exit(42)
	}
}

func SimpleConn(cobraSettings TaskSettings) (*tls.Conn, TaskSettings) {
	var settings TaskSettings
	if cobraSettings.NoRC {
		settings = cobraSettings
	} else {
		rc, err := ReadRC()
		if err != nil {
			fmt.Println("Warning: We couldn't find or open your taskrc file")
			settings = cobraSettings
		} else {
			settings = MakeSettings(rc)
		}
	}
	if settings.Server == "" {
		fmt.Println("You didn't specify a sever! Cowardly exiting.")
		os.Exit(69)
	}
	if settings.CACert == "" {
		fmt.Println("No CACert specified! Cowardly exiting.")
		os.Exit(69)
	}
	if settings.Certificate == "" {
		fmt.Println("No user cert specified for mtls")
		os.Exit(69)
	}
	if settings.Key == "" {
		fmt.Println("No user key specified for mtls")
		os.Exit(69)
	}
	conn := Connect(settings)
	return conn, settings
}

func Recv(conn *tls.Conn) []byte {
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
	headers["client"] = fmt.Sprintf("taskc-go %s", Version())
	headers["protocol"] = "v1"
	headers["org"] = org
	headers["key"] = uuid
	headers["user"] = user
	return headers
}

func Stats(conn *tls.Conn, credentials string) {
	msg := mkMessage(credentials)
	msg["type"] = "statistics"
	msgFinal := finalizeMessage(msg)
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	// log.Println(msgFinal)

	conn.Write(bytes.NewBufferString(msgFinal).Bytes())
}

func Pull(conn *tls.Conn, credentials string) {
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

func (t taskResponse) String() string {
	return fmt.Sprintf("%+v", t.Tasks)
}

type TaskSettings struct {
	Key, Server, Certificate, CACert, Creds string
	NoRC, Insecure                          bool
}

func ParseResponse(resp []byte) taskResponse {
	buff := bytes.NewBuffer(resp)
	scanner := bufio.NewScanner(buff)
	var tasks []task
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
	// log.Println(parsed)
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

func MakeSettings(rc map[string]string) TaskSettings {
	return TaskSettings{
		Server:      rc["taskd.server"],
		Certificate: rc["taskd.certificate"],
		Key:         rc["taskd.key"],
		CACert:      rc["taskd.ca"],
		Creds:       rc["taskd.credentials"],
	}
}

func Connect(settings TaskSettings) *tls.Conn {
	roots := x509.NewCertPool()
	cacert, err := ioutil.ReadFile(settings.CACert)
	if err != nil {
		panic("failed to load CA cert " + err.Error())
	}
	ok := roots.AppendCertsFromPEM(cacert)
	if !ok {
		panic("failed to parse root certificate")
	}

	cert, err := tls.LoadX509KeyPair(settings.Certificate, settings.Key)
	if err != nil {
		panic("failed to open cert/key: " + err.Error())
	}
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            roots,
		InsecureSkipVerify: settings.Insecure,
	}
	tlsConfig.BuildNameToCertificate()

	conn, err := tls.Dial("tcp", settings.Server, tlsConfig)
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	return conn
}
