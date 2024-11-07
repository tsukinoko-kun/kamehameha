package config

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"regexp"
)

type Probability float32

var probabilityRegexp = regexp.MustCompile(`^(\d+\.?\d*)\w*%$`)

func parseProbability(p string) (Probability, error) {
	var d string
	if m := probabilityRegexp.FindStringSubmatch(p); m != nil {
		d = m[1]
	} else {
		return 0, fmt.Errorf("invalid probability: %q", p)
	}

	var f float32
	if _, err := fmt.Sscanf(d, "%f", &f); err != nil {
		return 0, errors.Join(errors.New("failed to parse probability"), err)
	}

	return Probability(f / 100), nil
}

func (p Probability) Rand() bool {
	return rand.Float32() < float32(p)
}
