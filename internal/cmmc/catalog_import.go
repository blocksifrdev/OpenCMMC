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

// SourceCSV is an assisted import of NIST SP 800-171A June 2018 CSV data.
// NIST specifies that the PDF, not the CSV, is the authoritative source.
const SourceCSV = "https://csrc.nist.gov/files/pubs/sp/800/171/a/final/docs/sp800-171a-assessment-procedures.csv"

const SourcePDF = "https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-171a.pdf"

var reqToken = regexp.MustCompile(`\b3\.(?:[1-9]|1[0-4])\.[0-9]{1,2}\b`)
var objToken = regexp.MustCompile(`(?:3\.(?:[1-9]|1[0-4])\.[0-9]{1,2})?\[[a-z](?:[0-9]{1,2})?\]`)

type ImportSummary struct {
	Source                       string `json:"source"`
	SourceSHA256                 string `json:"source_sha256"`
	Authority                    string `json:"authority"`
	Requirements                 int    `json:"requirements"`
	Objectives                   int    `json:"objectives"`
	IndependentPDFReconciliation string `json:"independent_pdf_reconciliation"`
}

func csvHeaderIndex(row []string, want ...string) int {
	for i, s := range row {
		text := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(s, "\ufeff")))
		for _, w := range want {
			if strings.Contains(text, w) {
				return i
			}
		}
	}
	return -1
}

// Import171ACSV turns a locally supplied official-format NIST assessment CSV
// into a Level 2 workspace. The expected SHA-256 must be independently acquired
// and checked. The import proves structural completeness only, NEVER that the
// supplementary CSV was reconciled with the normative PDF.
func Import171ACSV(source []byte, scope, expectedSHA string) (ObjectiveWorkspace, ImportSummary, error) {
	if len(source) == 0 || len(source) > 12<<20 {
		return ObjectiveWorkspace{}, ImportSummary{}, errors.New("CSV must be 1..12 MiB")
	}
	sum := sha256.Sum256(source)
	digest := hex.EncodeToString(sum[:])
	if !hexDigest.MatchString(expectedSHA) || expectedSHA != digest {
		return ObjectiveWorkspace{}, ImportSummary{}, errors.New("source CSV SHA-256 must match independently verified --sha256")
	}
	r := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(source, []byte{0xef, 0xbb, 0xbf})))
	r.FieldsPerRecord = -1
	r.LazyQuotes = false
	rows, err := r.ReadAll()
	if err != nil {
		return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("parse CSV: %w", err)
	}
	if len(rows) < 2 || len(rows) > 15000 {
		return ObjectiveWorkspace{}, ImportSummary{}, errors.New("unexpected CSV row count")
	}
	header := rows[0]
	reqCol := csvHeaderIndex(header, "security requirement", "requirement id", "requirement number")
	objCol := csvHeaderIndex(header, "assessment objective", "determination statement")
	if reqCol < 0 || objCol < 0 || reqCol == objCol {
		return ObjectiveWorkspace{}, ImportSummary{}, errors.New("CSV must have separate security requirement and assessment objective columns; cannot safely infer format")
	}
	w, err := NewObjectiveWorkspace(2, scope)
	if err != nil {
		return ObjectiveWorkspace{}, ImportSummary{}, err
	}
	w.ProcedureSource = "NIST SP 800-171A (June 2018) supplemental CSV, SHA-256: " + digest + "; normative PDF reconciliation PENDING"
	w.ProcedureSHA256 = digest
	idx := map[string]int{}
	for i, c := range w.Controls {
		idx[c.ID] = i
	}
	known := map[string]map[string]string{}
	current := ""
	for rowN, row := range rows[1:] {
		if reqCol >= len(row) || objCol >= len(row) {
			return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("CSV row %d missing required columns", rowN+2)
		}
		rawReq := strings.TrimSpace(row[reqCol])
		if rawReq != "" {
			ids := reqToken.FindAllString(rawReq, -1)
			if len(ids) != 1 {
				return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("CSV row %d ambiguous requirement", rowN+2)
			}
			current = ids[0]
		}
		objText := strings.TrimSpace(row[objCol])
		if objText == "" {
			continue
		}
		if current == "" {
			return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("CSV row %d has objective before requirement", rowN+2)
		}
		ci, ok := idx[current]
		if !ok {
			return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("CSV row %d unknown requirement %q", rowN+2, current)
		}
		locations := objToken.FindAllStringIndex(objText, -1)
		if len(locations) == 0 {
			return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("CSV row %d objective text lacks bracketed ID", rowN+2)
		}
		if known[current] == nil {
			known[current] = map[string]string{}
		}
		for i, loc := range locations {
			token := objText[loc[0]:loc[1]]
			if strings.HasPrefix(token, "3.") && !strings.HasPrefix(token, current+"[") {
				return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("CSV row %d objective references wrong requirement %q", rowN+2, token)
			}
			id := current + token[strings.Index(token, "["):]
			end := len(objText)
			if i+1 < len(locations) {
				end = locations[i+1][0]
			}
			statement := strings.TrimSpace(objText[loc[1]:end])
			if statement == "" {
				return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("CSV row %d objective %s lacks text", rowN+2, id)
			}
			if _, exists := known[current][id]; exists {
				return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("duplicate objective %s (row %d)", id, rowN+2)
			}
			known[current][id] = statement
			w.Controls[ci].Objectives = append(w.Controls[ci].Objectives, ObjectiveItem{ID: id, Decision: "UNASSESSED", Statement: statement})
		}
	}
	count := 0
	for i := range w.Controls {
		if len(w.Controls[i].Objectives) == 0 {
			return ObjectiveWorkspace{}, ImportSummary{}, fmt.Errorf("incomplete CSV: no objectives for %s", w.Controls[i].ID)
		}
		sort.Slice(w.Controls[i].Objectives, func(a, b int) bool { return w.Controls[i].Objectives[a].ID < w.Controls[i].Objectives[b].ID })
		count += len(w.Controls[i].Objectives)
	}
	if err := ValidateObjectiveWorkspace(w); err != nil {
		return ObjectiveWorkspace{}, ImportSummary{}, err
	}
	return w, ImportSummary{Source: SourceCSV, SourceSHA256: digest, Authority: SourcePDF, Requirements: len(w.Controls), Objectives: count, IndependentPDFReconciliation: "NOT_PERFORMED"}, nil
}
