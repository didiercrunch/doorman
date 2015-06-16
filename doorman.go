package doorman

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	mathrand "math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/dchest/siphash"
	"github.com/didiercrunch/doorman/shared"
	"gopkg.in/mgo.v2/bson"
)

var _ = fmt.Print

var ONE *big.Rat = big.NewRat(1, 1)

var rand *mathrand.Rand

func initRandomSeed() {
	kindOfRandomSeed := mathrand.NewSource(time.Now().Unix())
	rand = mathrand.New(kindOfRandomSeed)
}

func init() {
	initRandomSeed()
}

func IsEqual(f1, f2 *big.Rat) bool {
	return f1.Cmp(f2) == 0
}

type Doorman struct {
	Id                  bson.ObjectId  // the id of the doorman
	LastChangeTimestamp int64          // an always increasing int that represent the last time the doorman has beed updated
	Probabilities       []*big.Rat     //  The probability of each cases.  The sum of probabilities needs to be one
	wg                  sync.WaitGroup // waitgroup for goroutine safety
}

func NewDoorman(id bson.ObjectId, Probabilities []*big.Rat) *Doorman {
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

	if !IsEqual(w.sum(wu.Probabilities), ONE) {
		return errors.New("the sum of probabilities cannot be different than 1")
	}
	w.Probabilities = wu.Probabilities
	log.Printf("Updated doorman %v with new probabilities %v with timestamp %v", wu.Id, wu.Probabilities, wu.Timestamp)
	return nil
}

func (w *Doorman) sum(prob []*big.Rat) *big.Rat {
	ret := big.NewRat(0, 1)
	for _, p := range prob {
		ret = new(big.Rat).Add(ret, p)
	}
	return ret
}

func (w *Doorman) Validate() error {
	if len(w.Probabilities) == 0 {
		return errors.New("not initiated")
	}
	s := w.sum(w.Probabilities)
	if !IsEqual(s, ONE) {
		return errors.New("The sum of probabilities is not one")
	}
	return nil
}

func (w *Doorman) GetCase(choosenRandomPosition *big.Rat) uint {
	w.wg.Wait()
	var prob = big.NewRat(0, 1)
	for i, p := range w.Probabilities {
		prob = new(big.Rat).Add(prob, p)
		if choosenRandomPosition.Cmp(prob) <= 0 {
			return uint(i)
		}
	}
	panic("cannot have a probability above 1")
}

func (w *Doorman) GenerateRandomProbabilityFromInteger(data uint64) *big.Rat {
	var ret float64 = 0
	for i := 0; data > 0; i++ {
		if data&1 == 1 {
			ret += 1.0 / math.Pow(2, float64(i+1))
		}
		data >>= 1
	}
	rat := new(big.Rat)
	return rat.SetFloat64(ret)
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
	r := rand.Float64()
	return w.GetCase(new(big.Rat).SetFloat64(r))
}
