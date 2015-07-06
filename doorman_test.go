package doorman

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/didiercrunch/doorman/shared"
)

const oid = "507f1f77bcf86cd799439011"

var epsilon = 0.0

var ZERO = big.NewRat(0, 1)

func newDoorman(probabilities []*big.Rat) *Doorman {
	id := base64.URLEncoding.EncodeToString(make([]byte, 16))
	if ret, err := New(id, probabilities); err != nil {
		panic(err)
	} else {
		return ret
	}
}

func assertIsEqual(t *testing.T, expected, received *big.Rat) {
	if expected.Cmp(received) != 0 {
		t.Error("received", received, "but expected", expected)
	}
}

func getProbs(probs ...string) []*big.Rat {
	ret := make([]*big.Rat, len(probs))
	var ok bool
	for i, p := range probs {
		if ret[i], ok = new(big.Rat).SetString(p); !ok {
			panic(p)
		}
	}
	return ret
}

func TestNew(t *testing.T) {
	notBase64 := "this is not url encoded base 64 /"
	if _, err := New(notBase64, getProbs("1/4", "3/4")); err == nil {
		t.Error()
	}
	base64OfAndArrayOfLength15 := "MTIzNDU2Nzg5MDEyMzQ1"
	if _, err := New(base64OfAndArrayOfLength15, getProbs("1/4", "3/4")); err == nil {
		t.Error()
	}
	goodId := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	if _, err := New(goodId, getProbs("1/4", "1/4")); err == nil {
		t.Error()
	}
}

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

	w = &Doorman{Probabilities: getProbs("2/4", "3/4")}
	if w.Validate().Error() != "The sum of probabilities is not one" {
		t.Error()
	}

	w = &Doorman{Probabilities: getProbs("1/4", "3/4")}
	if w.Validate() != nil {
		t.Error()
	}

	w = &Doorman{Probabilities: getProbs("100000001/400000000", "3/4")}
	if w.Validate() == nil {
		t.Error("even very small diff should be significative")
	}
}

func TestGetCase(t *testing.T) {
	w := newDoorman(getProbs("1/4", "2/4", "1/4"))
	if c := w.GetCase(ZERO); c != 0 {
		t.Error("expected 0 but received", c)
	}

	if c := w.GetCase(ONE); c != 2 {
		t.Error("expected 2 but received", c)
	}
}

func TestGetCaseCoroutineSafety(t *testing.T) {
	w := newDoorman(getProbs("1/4", "2/4", "1/4"))
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
	assertIsEqual(t, big.NewRat(1, 2), w.GenerateRandomProbabilityFromInteger(1))

	assertIsEqual(t, new(big.Rat).SetFloat64(0.5+0*0.25+1*0.125), w.GenerateRandomProbabilityFromInteger(5))

	assertIsEqual(t, ONE, w.GenerateRandomProbabilityFromInteger(math.MaxUint64))
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
	w := &Doorman{LastChangeTimestamp: 10, Id: oid}
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
	w := &Doorman{LastChangeTimestamp: 0, Id: oid}
	m := &shared.DoormanUpdater{Timestamp: 0, Probabilities: getProbs("1/2", "1/2"), Id: oid}
	w.Update(m)
	if len(w.Probabilities) != 0 {
		t.Error()
	}

	m = &shared.DoormanUpdater{Timestamp: 2, Probabilities: getProbs("1/2", "1/2"), Id: oid}
	if err := w.Update(m); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(w.Probabilities, getProbs("1/2", "1/2")) {
		t.Error("bad prob")
	}

	m = &shared.DoormanUpdater{Timestamp: 3, Probabilities: getProbs("1/2", "1/4"), Id: oid}
	if err := w.Update(m); err == nil {
		t.Error("should received an error")
	} else if !reflect.DeepEqual(w.Probabilities, getProbs("1/2", "1/2")) {
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

	wab := newDoorman(getProbs("1/3", "1/3", "1/3"))
	wab.Id = oid
	createMessage := func() string {
		msg := &shared.DoormanUpdater{Timestamp: 2, Probabilities: getProbs("1/2", "1/4", "1/4"), Id: oid}
		if ret, err := json.Marshal(msg); err != nil {
			panic(err)
		} else {
			return string(ret)
		}
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+wab.Id {
			t.Error("bad url", r.URL)
		}
		fmt.Fprint(w, createMessage())
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()
	if err := wab.UpdateHard(ts.URL); err != nil {
		t.Error(err)
	}
	if l := wab.Length(); l != 3 {
		t.Error(l)
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
	wab := newDoorman(getProbs("1/2", "1/2"))
	var sum = 0
	for i := 0; i < n; i++ {
		sum += int(wab.GetRandomCase())
	}

	if err := IsExtremeBinomialResult(sum, float64(n), p); err != nil {
		t.Error(err)
	}
}

func TestGetCaseFromString(t *testing.T) {
	w := newDoorman(getProbs("1/4", "1/2", "1/4"))
	expts := map[string]uint{
		"Հայաստան..":   2,
		"საქართველო":   1,
		"Azərbaycan..": 0,
	}
	for data, expt := range expts {
		if res := w.GetCaseFromString(data); res != expt {
			t.Error("incorrect results for: ", data, res)
		}
	}
}

func TestConsistencyOfDoormenWhenProbabilityChanges(t *testing.T) {
	w := "dddddddddddddddddddddd"
	for i := 1; i < 100; i++ {
		doorman := newDoorman(getProbs(strconv.Itoa(i)+"/100", strconv.Itoa(100-i)+"/100"))
		if doorman.GetCaseFromString(w) != 0 {
			t.Error()
		}
	}
}

// bellow test is not yet in doorman's specification.  It is not clear how
//we want to hash numbers because of the many number types in many programming
// language
func _TestGetCaseFromInt(t *testing.T) {
	w := newDoorman(getProbs("1/4", "1/2", "1/4"))
	expts := map[int]uint{
		2:   2,
		13:  2,
		195: 1,
	}
	for data, expt := range expts {
		if res := w.getCaseFromInt(data); res != expt {
			t.Error("incorrect results for: ", data, res)
		}
	}
}

func BenchmarkGetCaseFromData(b *testing.B) {
	var data = randomBytes(1024 * 1024)
	w := newDoorman(getProbs("10/100", "40/100", "40/100", "5/100", "5/100"))
	for i := 0; i < b.N; i++ {
		w.GetCaseFromData(data)
	}
}
