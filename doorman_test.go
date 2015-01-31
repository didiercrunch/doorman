package doorman

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/didiercrunch/doorman/shared"
	"gopkg.in/mgo.v2/bson"
)

func assertIsAbout(t *testing.T, expected, received float64) {
	if math.Abs(expected-received) > epsilon {
		t.Error("received", received, "but expected", expected)
	}
}

func assertIsNotAbout(t *testing.T, expected, received float64) {
	if math.Abs(expected-received) < epsilon {
		t.Error("received", received, "but expected something different than", expected)
	}
}

var oid = "507f1f77bcf86cd799439011"

func TestIsZero(t *testing.T) {
	if !IsZero(epsilon / 2) {
		t.Error()
	}

	if !IsZero(-epsilon / 2) {
		t.Error()
	}
	if IsZero(epsilon * 1.5) {
		t.Error()
	}
	if IsZero(-epsilon * 1.5) {
		t.Error()
	}
}

func TestIsEqual(t *testing.T) {
	if !IsEqual(1.0, 1.0+epsilon/2) {
		t.Error()
	}
	if IsEqual(1.0, 1.0+epsilon*1.5) {
		t.Error()
	}
}

func TestValidate(t *testing.T) {
	w := &Doorman{}
	if w.Validate().Error() != "not initiated" {
		t.Error()
	}

	w = &Doorman{Probabilities: []float64{0.5, 0.75}}
	if w.Validate().Error() != "The sum of probabilities is not one" {
		t.Error()
	}

	w = &Doorman{Probabilities: []float64{0.25, 0.75}}
	if w.Validate() != nil {
		t.Error()
	}

	w = &Doorman{Probabilities: []float64{0.2500000000000002, 0.75}}
	if w.Validate() != nil {
		t.Error()
	}
}

func TestGetCase(t *testing.T) {
	w := NewDoorman(bson.NewObjectId(), []float64{0.25, 0.5, 0.25})
	if c := w.GetCase(0); c != 0 {
		t.Error("expected 0 but received", c)
	}

	if c := w.GetCase(1); c != 2 {
		t.Error("expected 2 but received", c)
	}
}

func TestGetCaseCoroutineSafety(t *testing.T) {
	w := NewDoorman(bson.NewObjectId(), []float64{0.25, 0.5, 0.25})
	i := 0
	w.wg.Add(1)
	go func() {
		i++
		w.wg.Done()
	}()
	w.GetCase(0)
	if i != 1 {
		t.Error("error with goroutine safety")
	}
}

func TestGenerateRandomProbabilityFromBitSlice(t *testing.T) {
	w := new(Doorman)
	assertIsAbout(t, 0.5, w.GenerateRandomProbabilityFromInteger(1))

	assertIsAbout(t, 0.5+0*0.25+1*0.125, w.GenerateRandomProbabilityFromInteger(5))

	assertIsAbout(t, 1, w.GenerateRandomProbabilityFromInteger(math.MaxUint64))
}

func TestHash(t *testing.T) {
	m := []byte("doormen are great")
	w := new(Doorman)
	var expt uint64 = 0x3973fc1b3e324215
	result := w.Hash(m)

	if result != expt {
		t.Error(fmt.Sprintf("bad hashing, expected: %x, result: %x", expt, result))
	}
}

func TestUpdateTimestamp(t *testing.T) {

	w := &Doorman{LastChangeTimestamp: 10, Id: bson.ObjectIdHex(oid)}
	m := &shared.DoormanUpdater{Timestamp: 9}
	w.Update(m)
	if w.LastChangeTimestamp != 10 {
		t.Error()
	}

	m = &shared.DoormanUpdater{Timestamp: 11, Probabilities: []float64{}, Id: oid}
	w.Update(m)
	if w.LastChangeTimestamp != 11 {
		t.Error()
	}
}

func TestUpdateProbabilities(t *testing.T) {
	w := &Doorman{LastChangeTimestamp: 0, Id: bson.ObjectIdHex(oid)}
	m := &shared.DoormanUpdater{Timestamp: 0, Probabilities: []float64{0.5, 0.5}, Id: oid}
	w.Update(m)
	if len(w.Probabilities) != 0 {
		t.Error()
	}

	m = &shared.DoormanUpdater{Timestamp: 2, Probabilities: []float64{0.5, 0.5}, Id: oid}
	if err := w.Update(m); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(w.Probabilities, []float64{0.5, 0.5}) {
		t.Error("bad prob")
	}

	m = &shared.DoormanUpdater{Timestamp: 3, Probabilities: []float64{0.5, 0.25}, Id: oid}
	if err := w.Update(m); err == nil {
		t.Error("should received an error")
	} else if !reflect.DeepEqual(w.Probabilities, []float64{0.5, 0.5}) {
		t.Error("bad prob")
	}

}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i % 256)
	}
	return b
}

func TestUpdateHard(t *testing.T) {

	wab := NewDoorman(bson.ObjectIdHex(oid), []float64{})
	createMessage := func() string {
		msg := &shared.DoormanUpdater{Timestamp: 2, Probabilities: []float64{0.5, 0.25, 0.25}, Id: oid}
		if ret, err := json.Marshal(msg); err != nil {
			panic(err)
		} else {
			return string(ret)
		}
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+wab.Id.Hex() {
			t.Error("bad url", r.URL)
		}
		fmt.Fprint(w, createMessage())
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()
	if err := wab.UpdateHard(ts.URL); err != nil {
		t.Error(err)
	}
	if wab.Length() != 3 {
		t.Error()
	}
}

func TestGetRandomCase(t *testing.T) {
	wab := NewDoorman(bson.NewObjectId(), []float64{0.5, 0.5})
	var ret uint = 0
	for i := 0; i < 10000; i++ {
		ret += wab.GetRandomCase()
	}
	if ret < 4900 || ret > 5100 {
		t.Error("very improbable result of ", ret)
	}
}

func BenchmarkGetCaseFromData(b *testing.B) {
	var data = randomBytes(1024 * 1024)
	w := NewDoorman(bson.NewObjectId(), []float64{0.1, 0.4, 0.4, 0.05, 0.05})
	for i := 0; i < b.N; i++ {
		w.GetCaseFromData(data)
	}
}
