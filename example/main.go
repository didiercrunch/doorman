package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/didiercrunch/doorman"
	"gopkg.in/mgo.v2/bson"
)

var wab = doorman.NewDoorman(bson.ObjectIdHex("55048bfff6fb5e213d000001"), []float64{0.25, 0.5, 0.10, 0.05, 0.1})

func init() {
	if err := wab.NanoMsgSubscriber("tcp://127.0.0.1:40899"); err != nil {
		panic(err)
	}
}

func GetDoormanValue(r *http.Request) (string, int) {
	switch wab.GetCaseFromData([]byte(r.URL.String())) {
	case 0:
		return "blue", 0
	case 1:
		return "red", 1
	case 2:
		return "pink", 2
	case 3:
		return "yellow", 3
	case 4:
		return "magenta", 4
	}
	log.Panic("impossible value returned by doorman")
	return "", 0
}

var tmpl = `
<!DOCTYPE html>
<html>
	<head></head>
	<body style="background:{{ .color }};width:100%;height:100%">
		<h1 style="text-align:center;">the doorman value is {{ .value }}</h1>
		<p style="text-align:center;">
			this doorman depend solely on the url.  If you hit different url you
			will have different doorman results.
		</p>
	</body>
</html>
`

func Handler(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("foo").Parse(tmpl)
	if err != nil {
		panic(err)
	}
	m := make(map[string]interface{})
	m["color"], m["value"] = GetDoormanValue(r)
	if err := t.Execute(w, m); err != nil {
		panic(err)
	}
}

func getPort() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	} else {
		return "8080"
	}
}

func main() {
	http.HandleFunc("/", Handler)
	port := getPort()
	log.Println("serving on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
