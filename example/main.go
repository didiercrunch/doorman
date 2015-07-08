package main

import (
	"html/template"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/didiercrunch/doorman"
)

func getProbs(probs ...float64) []*big.Rat {
	ret := make([]*big.Rat, len(probs))
	for i, p := range probs {
		ret[i] = new(big.Rat).SetFloat64(p)
	}
	return ret
}

var wab *doorman.Doorman

func init() {
	var err error
	if wab, err = doorman.New("zST8_FWcPDT2-14wDQAAAQ==", getProbs(1, 0, 0)); err != nil {
		panic(err)
	}
	if err := wab.Subscriber("http://localhost:1999"); err != nil {
		panic(err)
	}
}

func GetDoormanValue(r *http.Request) (string, int) {
	switch wab.GetCaseFromData([]byte(r.URL.String())) {
	case 0:
		return "red", 0
	case 1:
		return "green", 1
	case 2:
		return "blue", 2
	}
	log.Panic("impossible value returned by doorman")
	return "", 0
}

var tmpl = `
<!DOCTYPE html>
<html>
	<head></head>
	<body style="background:{{ .color }};width:100%;height:100%">
		<div style="width: 62.5rem;">
		<h1 style="text-align:center;">the doorman value is {{ .value }}</h1>
		<p style="text-align:center;">
			This doorman depend solely on the url.  If you hit different url you
			will have different doorman results.  For example, here is a random
			url: <a href="{{ .nextUrl }}" style="color: black;"> {{ .nextUrl }} </a>
		</p>

		</div>

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
	m["nextUrl"] = "/" + strconv.Itoa(rand.Int())
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
