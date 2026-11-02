package cmmc

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// The June 2018 NIST PDF is authoritative; the supplemental CSV is an import aid.
const SourceCSV = "https://csrc.nist.gov/files/pubs/sp/800/171/a/final/docs/sp800-171a-assessment-procedures.csv"
const SourcePDF = "https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-171a.pdf"

var csvIdentifier = regexp.MustCompile(`^(3\.(?:[1-9]|1[0-4])\.[0-9]{1,2})(\[[a-z][0-9]{0,2}\])?$`)

type ImportSummary struct {
	Source                       string `json:"source"`
	SourceSHA256                 string `json:"source_sha256"`
	Authority                    string `json:"authority"`
	Requirements                 int    `json:"requirements"`
	Objectives                   int    `json:"objectives"`
	IndependentPDFReconciliation string `json:"independent_pdf_reconciliation"`
}

func csvHeaderIndex(row []string, wants ...string) int {
	for i, s := range row {
		lower := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(s, "\ufeff")))
		for _, w := range wants {
			if strings.Contains(lower, w) {
				return i
			}
		}
	}
	return -1
}

// Import171ACSV supports the actual official NIST spreadsheet: each row
// identifies either requirement 3.x.y (including methods) or one objective
// 3.x.y[a]. Refuses partial coverage and never claims PDF reconciliation.
func Import171ACSV(source []byte, scope, expectedSHA string) (ObjectiveWorkspace, ImportSummary, error) {
	if len(source) == 0 || len(source) > 12<<20 {
		return ObjectiveWorkspace{}, ImportSummary{}, errors.New("CSV must be 1..12 MiB")
	}
	sum := sha256.Sum256(source)
	digest := hex.EncodeToString(sum[:])
	if !hexDigest.MatchString(expectedSHA) || expectedSHA != digest {
		return ObjectiveWorkspace{}, ImportSummary{}, errors.New("source CSV SHA-256 must match independently verified --sha256")
	}
	reader := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(source, []byte{0xef, 0xbb, 0xbf})))
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("parse CSV: %w", err)
	}
	if len(rows) < 2 || len(rows) > 15000 {
		return ObjectiveWorkspace{}, ImportSummary{}, errors.New("unexpected CSV row count")
	}
	header := rows[0]
	idCol := csvHeaderIndex(header, "identifier")
	reqCol := csvHeaderIndex(header, "security requirement")
	objCol := csvHeaderIndex(header, "assessment objective")
	examineCol := csvHeaderIndex(header, "method and objects: examine")
	interviewCol := csvHeaderIndex(header, "method and objects: interview")
	testCol := csvHeaderIndex(header, "method and objects: test")
	if idCol < 0 || reqCol < 0 || objCol < 0 || len(map[int]bool{idCol: true, reqCol: true, objCol: true}) != 3 {
		return ObjectiveWorkspace{}, ImportSummary{}, errors.New("unsupported NIST CSV: Identifier, Security Requirement and Assessment Objective columns required")
	}
	if examineCol < 0 || interviewCol < 0 || testCol < 0 {
		return ObjectiveWorkspace{}, ImportSummary{}, errors.New("NIST assessment methods columns missing")
	}
	w, err := NewObjectiveWorkspace(2, scope)
	if err != nil {
		return ObjectiveWorkspace{}, ImportSummary{}, err
	}
	w.ProcedureSource = "NIST SP 800-171A (June 2018) supplemental CSV; SHA-256 " + digest + "; normative PDF reconciliation PENDING"
	w.ProcedureSHA256 = digest
	index := map[string]int{}
	for i, c := range w.Controls {
		index[c.ID] = i
	}
	requirementRow := map[string]bool{}
	objectives := map[string]bool{}
	count := 0
	for rowN, row := range rows[1:] {
		if idCol >= len(row) || reqCol >= len(row) || objCol >= len(row) || testCol >= len(row) {
			return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("CSV row %d missing required columns", rowN+2)
		}
		raw := strings.TrimSpace(row[idCol])
		if raw == "" {
			return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("CSV row %d missing identifier", rowN+2)
		}
		match := csvIdentifier.FindStringSubmatch(raw)
		if match == nil {
			return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("CSV row %d invalid identifier %q", rowN+2, raw)
		}
		id, suffix := match[1], match[2]
		ci, ok := index[id]
		if !ok {
			return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("CSV row %d unexpected requirement %q", rowN+2, id)
		}
		text := strings.TrimSpace(row[objCol])
		if suffix == "" {
			if requirementRow[id] {
				return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("duplicate requirement header %s", id)
			}
			requirementRow[id] = true
			if !strings.HasPrefix(strings.ToLower(text), "determine if") || strings.TrimSpace(row[reqCol]) == "" {
				return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("CSV row %d requirement header malformed", rowN+2)
			}
			w.Controls[ci].RequirementText = strings.TrimSpace(row[reqCol])
			w.Controls[ci].Methods = AssessmentMethodGuidance{Examine: strings.TrimSpace(row[examineCol]), Interview: strings.TrimSpace(row[interviewCol]), Test: strings.TrimSpace(row[testCol])}
			continue
		}
		if !requirementRow[id] {
			return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("CSV row %d objective before requirement header %q", rowN+2, id)
		}
		oid := id + suffix
		if objectives[oid] {
			return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("duplicate objective %s", oid)
		}
		if text == "" || len(text) > 4096 {
			return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("empty/oversized objective text for %s", oid)
		}
		objectives[oid] = true
		w.Controls[ci].Objectives = append(w.Controls[ci].Objectives, ObjectiveItem{ID: oid, Statement: text, Decision: "UNASSESSED"})
		count++
	}
	for i := range w.Controls {
		c := &w.Controls[i]
		if !requirementRow[c.ID] || len(c.Objectives) == 0 {
			return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("incomplete CSV: %s is missing a header or objectives", c.ID)
		}
		sort.Slice(c.Objectives, func(i, j int) bool { return c.Objectives[i].ID < c.Objectives[j].ID })
		for idx, obj := range c.Objectives {
			expected := fmt.Sprintf("%s[%c]", c.ID, 'a'+idx)
			if idx >= 26 || obj.ID != expected {
				return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("non-contiguous objective IDs in %s: expected %s, found %s", c.ID, expected, obj.ID)
			}
		}
	}
	if err := ValidateObjectiveWorkspace(w); err != nil {
		return ObjectiveWorkspace{}, ImportSummary{}, err
	}
	return w, ImportSummary{Source: SourceCSV, SourceSHA256: digest, Authority: SourcePDF, Requirements: len(w.Controls), Objectives: count, IndependentPDFReconciliation: "NOT_PERFORMED"}, nil
}
