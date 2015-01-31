package doorman

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	mathrand "math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/dchest/siphash"
	"github.com/didiercrunch/doorman/shared"
	"gopkg.in/mgo.v2/bson"
)

// the smallest value that makes two floats the same.  Consequently x == y
// if |x - y| < epsilon.  It is a known issue that one can find three numbers
// a,b,c such that a == b, b == c and a != c. Hence, it is the programmer role
// to be very careful about equalities and inequalities.
const epsilon float64 = 0.000000001

var rand *mathrand.Rand

func initRandomSeed() {
	kindOfRandomSeed := mathrand.NewSource(time.Now().Unix())
	rand = mathrand.New(kindOfRandomSeed)
}

func init() {
	initRandomSeed()
}

func IsZero(f float64) bool {
	if math.Abs(f) > epsilon {
		return false
	}
	return true
}

func IsEqual(f1, f2 float64) bool {
	return IsZero(f1 - f2)
}

type Doorman struct {
	Id                  bson.ObjectId  // the id of the doorman
	LastChangeTimestamp int64          // an always increasing int that represent the last time the doorman has beed updated
	Probabilities       []float64      //  The probability of each cases.  The sum of probabilities needs to be one
	wg                  sync.WaitGroup // waitgroup for goroutine safety
}

func NewDoorman(id bson.ObjectId, Probabilities []float64) *Doorman {
	wab := &Doorman{}
	wab.Id = id
	wab.Probabilities = Probabilities
	return wab
}

func (w *Doorman) Length() int {
	return len(w.Probabilities)
}

func (w *Doorman) UpdateHard(baseURL string) error {
	r, err := http.Get(baseURL + "/" + w.Id.Hex())
	if err != nil {
		return err
	}
	message := new(shared.DoormanUpdater)
	d := json.NewDecoder(r.Body)
	if err := d.Decode(message); err != nil {
		return err
	} else {
		return w.Update(message)
	}
}

func (w *Doorman) Update(wu *shared.DoormanUpdater) error {
	if wu.Timestamp <= w.LastChangeTimestamp {
		return nil
	}
	if wu.Id != w.Id.Hex() {
		return errors.New("bad doorman id")
	}
	w.wg.Add(1)
	defer w.wg.Done()
	w.LastChangeTimestamp = wu.Timestamp

	if !IsEqual(w.sum(wu.Probabilities), 1) {
		return errors.New("the sum of probabilities cannot be different than 1")
	}
	w.Probabilities = wu.Probabilities
	log.Printf("Updated doorman %v with new probabilities %v with timestamp %v", wu.Id, wu.Probabilities, wu.Timestamp)
	return nil
}

func (w *Doorman) sum(prob []float64) float64 {
	var ret float64 = 0
	for _, p := range prob {
		ret += p
	}
	return ret
}

func (w *Doorman) Validate() error {
	if len(w.Probabilities) == 0 {
		return errors.New("not initiated")
	}
	if !IsZero(w.sum(w.Probabilities) - 1.0) {
		return errors.New("The sum of probabilities is not one")
	}
	return nil
}

func (w *Doorman) GetCase(choosenRandomPosition float64) uint {
	w.wg.Wait()
	var prob float64 = 0
	for i, p := range w.Probabilities {
		prob += p
		if choosenRandomPosition-epsilon < prob {
			return uint(i)
		}
	}
	panic("cannot have a probability above 1")
}

func (w *Doorman) GenerateRandomProbabilityFromInteger(data uint64) float64 {
	var ret float64 = 0
	for i := 0; data > 0; i++ {
		if data&1 == 1 {
			ret += 1.0 / math.Pow(2, float64(i+1))
		}
		data >>= 1
	}
	return ret
}

func (w *Doorman) Hash(data ...[]byte) uint64 {
	h := siphash.New(make([]byte, 16))
	for _, datum := range data {
		h.Write(datum)
	}
	return h.Sum64()
}

func (w *Doorman) GetCaseFromData(data ...[]byte) uint {
	random := w.GenerateRandomProbabilityFromInteger(w.Hash(data...))
	return w.GetCase(random)
}

func (w *Doorman) GetRandomCase() uint {
	return w.GetCase(rand.Float64())
}
