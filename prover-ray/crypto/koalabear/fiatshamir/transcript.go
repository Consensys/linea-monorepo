// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package fiatshamir implements a Fiat-Shamir transcript for non-interactive
// proof systems.
//
// A [Transcript] derives verifier challenges deterministically from prover
// messages using a hash function, turning an interactive protocol into a
// non-interactive one.
//
// Challenges are registered by name (at construction or via [Transcript.NewChallenge]),
// bound to values with [Transcript.Bind], and computed sequentially with
// [Transcript.ComputeChallenge].
package fiatshamir

import (
	"errors"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/linea-monorepo/prover-ray/crypto/koalabear/hash"
)

// errChallengeNotFound is returned when a wrong challenge name is provided.
var (
	errChallengeNotFound            = errors.New("challenge not recorded in the transcript")
	errChallengeAlreadyComputed     = errors.New("challenge already computed, cannot be binded to other values")
	errPreviousChallengeNotComputed = errors.New("the previous challenge is needed and has not been computed")
	errChallengeAlreadyExists       = errors.New("this challenge name is already used and recorded")
)

const (
	challengeIDDomainTag uint64 = 0x46534944 // "FSID"
)

// Transcript implements a Fiat-Shamir transcript for transforming
// an interactive protocol into a non-interactive one.
// Challenges must be computed in the order they were registered.
type Transcript struct {
	// hash function that is used.
	h hash.FieldHasher

	challenges         []challenge // the order matters
	nameToChallengePos map[string]int
}

type challenge struct {
	bindings   []koalabear.Element // bindings stores the variables a challenge is binded to.
	name       string
	value      hash.Digest // value stores the computed challenge
	isComputed bool
}

// NewTranscript creates a new Fiat-Shamir transcript using the given hash function.
// challengesID are the names of the challenges that will be computed; the order
// matters, as each challenge depends on the previous one.
// Additional challenges can be appended later with [Transcript.NewChallenge].
//
// It panics if duplicate challenge names are provided.
func NewTranscript(h hash.FieldHasher, challengesID ...string) *Transcript {
	t := &Transcript{
		challenges:         make([]challenge, 0, len(challengesID)),
		nameToChallengePos: make(map[string]int, len(challengesID)),
		h:                  h,
	}
	for _, id := range challengesID {
		if _, ok := t.nameToChallengePos[id]; ok {
			panic("duplicate challenge name: " + id)
		}
		t.nameToChallengePos[id] = len(t.challenges)
		t.challenges = append(t.challenges, challenge{name: id})
	}
	return t
}

// Bind binds a value to the given challenge. A challenge can be bound to an
// arbitrary number of values, but the order in which the values are added
// matters. It returns an error if the challenge does not exist or has already
// been computed.
func (t *Transcript) Bind(challengeID string, bValue []koalabear.Element) error {

	pos, ok := t.nameToChallengePos[challengeID]
	if !ok {
		return errChallengeNotFound
	}

	currentChallenge := t.challenges[pos]
	if currentChallenge.isComputed {
		return errChallengeAlreadyComputed
	}

	bCopy := make([]koalabear.Element, len(bValue))
	copy(bCopy, bValue)
	currentChallenge.bindings = append(currentChallenge.bindings, bCopy...)
	t.challenges[pos] = currentChallenge

	return nil
}

// NewChallenge appends a new challenge to the transcript. The newly added
// challenge becomes the last in the computation order. It returns an error if a
// challenge with the same name already exists.
func (t *Transcript) NewChallenge(challengeID string) error {
	if _, ok := t.nameToChallengePos[challengeID]; ok {
		return errChallengeAlreadyExists
	}
	nbChallenges := len(t.challenges)
	challenge := challenge{
		name:       challengeID,
		isComputed: false,
	}
	t.challenges = append(t.challenges, challenge)
	t.nameToChallengePos[challengeID] = nbChallenges
	return nil
}

// ComputeChallenge computes and returns the challenge corresponding to the
// given name. Challenges must be computed sequentially in the order they were
// registered. The result is:
//   - H(name || bound_values...) for the first challenge
//   - H(name || previous_challenge || bound_values...) for subsequent challenges
//
// If the challenge has already been computed, the cached value is returned.
// It returns an error if the challenge does not exist or if the previous
// challenge has not been computed yet.
func (t *Transcript) ComputeChallenge(challengeID string) ([8]koalabear.Element, error) {

	pos, ok := t.nameToChallengePos[challengeID]
	if !ok {
		return [8]koalabear.Element{}, errChallengeNotFound
	}

	// if the challenge was already computed we return it
	challenge := t.challenges[pos]
	if challenge.isComputed {
		return challenge.value, nil
	}

	// reset before populating the internal state
	t.h.Reset()
	defer t.h.Reset()

	t.h.WriteElements(hash.StringToElements(challengeIDDomainTag, challengeID)...)

	// write the previous challenge if it's not the first challenge
	if pos != 0 {
		if !t.challenges[pos-1].isComputed {
			return [8]koalabear.Element{}, errPreviousChallengeNotComputed
		}
		t.h.WriteElements(t.challenges[pos-1].value[:]...)
	}

	// write the binded values in the order they were added
	for _, b := range challenge.bindings {
		t.h.WriteElements(b)
	}

	// compute the hash of the accumulated values
	challenge.value = t.h.Sum()
	challenge.isComputed = true

	t.challenges[pos] = challenge

	return challenge.value, nil

}
