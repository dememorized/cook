package conversion

import (
	"math/big"
	"regexp"
	"strconv"
)

type Numeral string

var rationalNumberRegex = regexp.MustCompile("([0-9]+)\\s*/\\s*([0-9]+)")

func (n Numeral) Valid() bool {
	_, err := n.Float()
	return err == nil
}

func (n Numeral) Float() (float64, error) {
	if rat, err := n.rational(); err == nil {
		val, _ := rat.Float64()
		return val, nil
	}

	return strconv.ParseFloat(string(n), 64)
}

func (n Numeral) Rational() (*big.Rat, error) {
	if rat, err := n.rational(); err == nil {
		return rat, nil
	}

	f, err := strconv.ParseFloat(string(n), 64)
	if err != nil {
		return nil, err
	}
	rat, _ := big.NewFloat(f).Rat(nil)
	return rat, nil
}

func (n Numeral) rational() (*big.Rat, error) {
	sub := rationalNumberRegex.FindStringSubmatch(string(n))
	if sub == nil {
		return nil, big.ErrNaN{}
	}

	a, err := strconv.ParseInt(sub[1], 10, 64)
	if err != nil {
		return nil, err
	}
	b, err := strconv.ParseInt(sub[2], 10, 64)
	if err != nil {
		return nil, err
	}

	return big.NewRat(a, b), nil
}
