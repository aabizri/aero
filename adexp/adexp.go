/*
Package adexp provides marshalling and unmarshalling of ADEXP documents.
See: http://www.eurocontrol.int/sites/default/files/publication/files/20111001-adexp-spec-v3.1.pdf

Architecture:
	lexer provides the definition of a ADEXP lexer (DONE)
	parser provides the definition of a ADEXP parser (WIP)
	serialiser provides the definition of an ADEXP serialiser (TODO)
	this package wraps it all together so as to provide easy & simple unmarshalling of ADEXP documents (WIP)
*/
package adexp

import (
	"errors"
)

// ADEXPMsg is free text conforming to the syntax described for an ADEXP message
type ADEXPMsg string

// AircraftID identifies an aircraft
// 2{ALPHANUM}7
type AircraftID string

func (id AircraftID) check() error {
	// First check the length
	if !(2 <= len(id) && len(id) <= 7) {
		return errors.New("Invalid length for AircraftID")
	}
	// Now check if its correct ALPHANUM
	// TODO

	return nil
}

// ATFMRDYState is the ready status of the flight
// It can be either 'D' or 'N'.
type ATFMRDYState string

const (
	// ATFMRDYStateReady codes for Ready to depart
	ATFMRDYStateReady ATFMRDYState = "D"
	// ATFMRDYStateNotReady codes for Not Ready to depart
	ATFMRDYStateNotReady = "N"
)

// TimeHHMM codes for a departure hour/minute
type TimeHHMM string

// Date is a date in YYMMDD format
type Date string

// FieldID is a valid ADEXP field name
type FieldID string

// ADEXP is a Message in a ADEXP (Air traffic services Data EXchange Presentation) format.
type ADEXP struct {
	// The TITLE field is always the first field of an ADEXP message.
	// It has to be all-uppercase IA-5 alphabet.
	TITLE string

	// ADEP is the ICAO location indicator of the aerodrome of departure
	// or the indication AFIL meaning an air-filed flight plan
	// or "ZZZZ" when no ICAO location indicator is assigned to the aerodrome of departure
	ADEP string

	// ADES is the ICAO location indicator of the aerodrome of destination
	// or "ZZZZ" when no ICAO location indicator is assigned to the aerodrome of destination
	ADES string

	// ARCID is the Aircraft Identification.
	// May be the registration marking of the Aircraft,
	// or the ICAO designator of the Aircraft Operator followed by the Flight Identifier
	ARCID AircraftID

	// CTOT is the Calculated Take-Off Time: reference time of an ATFM slot.
	CTOT TimeHHMM

	// EOBD is the estimated Off-Block Date.
	EOBD Date

	// EOBT is the estimated Off-Block Time.
	EOBT TimeHHMM

	// ERRFIELD is the Adexp Name of erroneous fields.
	ERRFIELD FieldID

	// NEWCTOT is a new Calculated Take-Off time, as updated by ETFMS
	NEWCTOT TimeHHMM
}

func (msg ADEXP) MarshalText() ([]byte, error) {
	return nil, nil
}

func (msg ADEXP) MarshalTextCustom(separator string, newlines bool, indentation bool) ([]byte, error) {
	return nil, nil
}
