package doorman

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/didiercrunch/doorman/shared"
	"gopkg.in/mgo.v2/bson"
)

var epsilon = 0.0
var ZERO = big.NewRat(0, 1)

func assertIsAbout(t *testing.T, expected, received *big.Rat) {
	if expected.Cmp(received) != 0 {
		t.Error("received", received, "but expected", expected)
	}
}

func assertIsNotAbout(t *testing.T, expected, received *big.Rat) {
	if expected.Cmp(received) != 0 {
		t.Error("received", received, "but expected something different than", expected)
	}
}

func getProbs(probs ...float64) []*big.Rat {
	ret := make([]*big.Rat, len(probs))
	for i, p := range probs {
		ret[i], _ = new(big.Rat).SetString(fmt.Sprint(p))
	}
	return ret
}

var oid = "507f1f77bcf86cd799439011"

func TestIsEqual(t *testing.T) {
	if !IsEqual(big.NewRat(2, 2), big.NewRat(1, 1)) {
		t.Error()
	}
	if IsEqual(big.NewRat(2, 2), big.NewRat(2, 1)) {
		t.Error()
	}
}

func TestValidate(t *testing.T) {
	w := &Doorman{}
	if w.Validate().Error() != "not initiated" {
		t.Error()
	}

	w = &Doorman{Probabilities: getProbs(0.5, 0.75)}
	if w.Validate().Error() != "The sum of probabilities is not one" {
		t.Error()
	}

	w = &Doorman{Probabilities: getProbs(0.25, 0.75)}
	if w.Validate() != nil {
		t.Error()
	}

	w = &Doorman{Probabilities: getProbs(0.2500000000002, 0.75)}
	if w.Validate() == nil {
		t.Error("even very small diff should be significative")
	}
}

func TestGetCase(t *testing.T) {
	w := NewDoorman(bson.NewObjectId(), getProbs(0.25, 0.5, 0.25))
	if c := w.GetCase(ZERO); c != 0 {
		t.Error("expected 0 but received", c)
	}

	if c := w.GetCase(ONE); c != 2 {
		t.Error("expected 2 but received", c)
	}
}

func TestGetCaseCoroutineSafety(t *testing.T) {
	w := NewDoorman(bson.NewObjectId(), getProbs(0.25, 0.5, 0.25))
	i := 0
	w.wg.Add(1)
	go func() {
		i++
		w.wg.Done()
	}()
	w.GetCase(ZERO)
	if i != 1 {
		t.Error("error with goroutine safety")
	}
}

func TestGenerateRandomProbabilityFromBitSlice(t *testing.T) {
	w := new(Doorman)
	assertIsAbout(t, big.NewRat(1, 2), w.GenerateRandomProbabilityFromInteger(1))

	//	assertIsAbout(t, 0.5+0*0.25+1*0.125, w.GenerateRandomProbabilityFromInteger(5))

	assertIsAbout(t, ONE, w.GenerateRandomProbabilityFromInteger(math.MaxUint64))
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

	m = &shared.DoormanUpdater{Timestamp: 11, Probabilities: getProbs(), Id: oid}
	w.Update(m)
	if w.LastChangeTimestamp != 11 {
		t.Error()
	}
}

func TestUpdateProbabilities(t *testing.T) {
	w := &Doorman{LastChangeTimestamp: 0, Id: bson.ObjectIdHex(oid)}
	m := &shared.DoormanUpdater{Timestamp: 0, Probabilities: getProbs(0.5, 0.5), Id: oid}
	w.Update(m)
	if len(w.Probabilities) != 0 {
		t.Error()
	}

	m = &shared.DoormanUpdater{Timestamp: 2, Probabilities: getProbs(0.5, 0.5), Id: oid}
	if err := w.Update(m); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(w.Probabilities, getProbs(0.5, 0.5)) {
		t.Error("bad prob")
	}

	m = &shared.DoormanUpdater{Timestamp: 3, Probabilities: getProbs(0.5, 0.25), Id: oid}
	if err := w.Update(m); err == nil {
		t.Error("should received an error")
	} else if !reflect.DeepEqual(w.Probabilities, getProbs(0.5, 0.5)) {
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

	wab := NewDoorman(bson.ObjectIdHex(oid), getProbs())
	createMessage := func() string {
		msg := &shared.DoormanUpdater{Timestamp: 2, Probabilities: getProbs(0.5, 0.25, 0.25), Id: oid}
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

func IsExtremeBinomialResult(x int, n, p float64) error {
	std := math.Sqrt(n * p * (1 - p))
	mean := n * p
	xf := float64(x)
	if minVal := mean - 3.5*std; xf < minVal {
		return errors.New(fmt.Sprintf("value too small of %v when minimal threshold is %v", x, minVal))
	}
	if maxVal := mean + 3.5*std; xf > maxVal {
		return errors.New(fmt.Sprintf("value too big of %v when maximum threshold is %v", x, maxVal))
	}
	return nil

}

func TestGetRandomCase(t *testing.T) {
	p := 0.5
	n := 10000
	wab := NewDoorman(bson.NewObjectId(), getProbs(p, 1-p))
	var sum = 0
	for i := 0; i < n; i++ {
		sum += int(wab.GetRandomCase())
	}

	if err := IsExtremeBinomialResult(sum, float64(n), p); err != nil {
		t.Error(err)
	}
}

func BenchmarkGetCaseFromData(b *testing.B) {
	var data = randomBytes(1024 * 1024)
	w := NewDoorman(bson.NewObjectId(), getProbs(0.1, 0.4, 0.4, 0.05, 0.05))
	for i := 0; i < b.N; i++ {
		w.GetCaseFromData(data)
	}
}
