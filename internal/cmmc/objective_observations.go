package cmmc

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ObjectiveObservation is a human-attested observation of a specific assessment
// objective using EXAMINE, INTERVIEW or TEST. A claim is not verified merely
// because a screenshot, log, or output file exists or has a matching SHA-256.
type ObjectiveObservation struct {
	ID           string `json:"id"`
	Method       string `json:"method"`
	Outcome      string `json:"outcome"` // SUPPORTS, CONTRADICTS, INCONCLUSIVE
	ObservedAt   string `json:"observed_at"`
	Observer     string `json:"observer"`
	EvidencePath string `json:"evidence_path"`
	Rationale    string `json:"rationale"`
}

var observationIdentifier = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._:-]{0,63}$`)

func ValidateObjectiveObservations(item ObjectiveItem) error {
	if len(item.Observations) > 32 {
		return errors.New("too many observations for objective")
	}
	seen := map[string]bool{}
	evidence := map[string]bool{}
	for _, e := range item.Evidence {
		evidence[e.Path] = true
	}
	for _, o := range item.Observations {
		if !observationIdentifier.MatchString(o.ID) || seen[o.ID] {
			return fmt.Errorf("invalid or duplicated observation identifier for %s", item.ID)
		}
		seen[o.ID] = true
		switch o.Method {
		case "EXAMINE", "INTERVIEW", "TEST":
		default:
			return fmt.Errorf("unsupported observation method %q", o.Method)
		}
		switch o.Outcome {
		case "SUPPORTS", "CONTRADICTS", "INCONCLUSIVE":
		default:
			return fmt.Errorf("unsupported observation outcome %q", o.Outcome)
		}
		if strings.TrimSpace(o.Observer) == "" || len(o.Observer) > 256 ||
			strings.TrimSpace(o.Rationale) == "" || len(o.Rationale) > 2048 {
			return fmt.Errorf("observation %s must identify reviewer and rationale", o.ID)
		}
		if len(o.EvidencePath) == 0 || len(o.EvidencePath) > 4096 || !evidence[o.EvidencePath] {
			return fmt.Errorf("observation %s must reference an attached evidence file", o.ID)
		}
		if _, err := time.Parse(time.RFC3339, o.ObservedAt); err != nil {
			return fmt.Errorf("observation %s has invalid observed_at: %w", o.ID, err)
		}
	}
	return nil
}

// ReviewObjectiveObservations checks the observation time, policy freshness,
// method and evidence digest linkage. These are necessary conditions for
// reviewer-ready findings, never an independent effectiveness determination.
func ReviewObjectiveObservations(item ObjectiveItem, verifiedFiles map[string]bool, at time.Time, maxAgeDays int) (supports, contradicts bool, unresolved []string) {
	for _, obs := range item.Observations {
		observed, err := time.Parse(time.RFC3339, obs.ObservedAt)
		if err != nil || observed.After(at.Add(5*time.Minute)) ||
			at.Sub(observed) > time.Duration(maxAgeDays)*24*time.Hour {
			unresolved = append(unresolved, obs.ID+":STALE_OR_INVALID_OBSERVATION")
			continue
		}
		if !verifiedFiles[obs.EvidencePath] {
			unresolved = append(unresolved, obs.ID+":EVIDENCE_NOT_VERIFIED")
			continue
		}
		switch obs.Outcome {
		case "CONTRADICTS":
			contradicts = true // Override any positive reviewer assertion.
		case "SUPPORTS":
			if obs.Method == item.Method {
				supports = true
			} else {
				unresolved = append(unresolved, obs.ID+":METHOD_MISMATCH")
			}
		case "INCONCLUSIVE":
			unresolved = append(unresolved, obs.ID+":INCONCLUSIVE")
		}
	}
	return supports, contradicts, unresolved
}
