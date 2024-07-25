package wizard

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strconv"
)

// IOPStats collects general informations about the polynomial IOP. Ventilated by
// status, tag
type IOPStats struct {
	Records []Record
	AllTags []string
}

type Record struct {
	FullName      string
	Type          string
	Size          int
	Round         int
	Degree        int
	Visibility    string
	TagIndicators []bool
}

func (comp *CompiledIOP) Stats() *IOPStats {

	var (
		records      = []Record{}
		tags, tagMap = comp.allTags()
	)

	tagRow := func(obj interface{ Tags() []string }) []bool {
		var (
			row   = make([]bool, len(tags))
			cTags = obj.Tags()
		)

		for _, t := range cTags {
			pos := tagMap[t]
			row[pos] = true
		}

		return row
	}

	for _, c := range comp.coins.all() {

		rec := Record{
			FullName:      c.String(),
			Type:          reflect.TypeOf(c).String(),
			Round:         c.Round(),
			TagIndicators: tagRow(c),
		}

		records = append(records, rec)
	}

	for _, c := range comp.columns.all() {

		rec := Record{
			FullName:      c.String(),
			Type:          reflect.TypeOf(c).String(),
			Size:          c.Size(),
			Round:         c.Round(),
			Visibility:    c.Visibility().String(),
			TagIndicators: tagRow(&c),
		}

		records = append(records, rec)
	}

	for _, c := range comp.queries.all() {

		rec := Record{
			FullName:      c.String(),
			Type:          reflect.TypeOf(c).String(),
			Round:         c.Round(),
			TagIndicators: tagRow(c),
		}

		if glob, ok := c.(*QueryGlobal); ok {
			rec.Size = glob.domainSize
			rec.Degree = glob.Expr.board.Degree(func(m interface{}) int {
				if _, ok := m.(Column); ok {
					return 1
				}
				return 0
			})
		}

		records = append(records, rec)
	}

	return &IOPStats{
		Records: records,
		AllTags: tags,
	}
}

func (is *IOPStats) SaveAsCsv(f io.Writer) error {

	w := csv.NewWriter(f)

	header := []string{"fullName", "type", "size", "round", "degree", "visibility"}
	for _, t := range is.AllTags {
		header = append(header, "tag:"+t)
	}

	err := w.Write(header)
	if err != nil {
		return fmt.Errorf("could not write header as CSV: %w", err)
	}

	for _, rec := range is.Records {
		row := make([]string, 0, len(header))
		row = append(row,
			rec.FullName,
			rec.Type,
			intToStringOrEmpty(rec.Size),
			intToStringOrEmpty(rec.Round),
			intToStringOrEmpty(rec.Degree),
			rec.Visibility,
		)

		for i := range rec.TagIndicators {
			row = append(row, boolToStringOrEmpty(rec.TagIndicators[i]))
		}

		err := w.Write(row)

		if err != nil {
			return fmt.Errorf("could not write header as CSV: %w", err)
		}
	}

	w.Flush()

	if w.Error() != nil {
		return fmt.Errorf("could not flush CSV writer: %w", err)
	}

	return nil
}

func (comp *CompiledIOP) allTags() ([]string, map[string]int) {

	var (
		tagSet = map[string]struct{}{}
		tagMap = map[string]int{}
	)

	for _, c := range comp.coins.all() {
		cTags := c.Tags()
		for _, t := range cTags {
			tagSet[t] = struct{}{}
		}
	}

	for _, q := range comp.queries.all() {
		qTags := q.Tags()
		for _, t := range qTags {
			tagSet[t] = struct{}{}
		}
	}

	for _, c := range comp.columns.all() {
		cTags := c.Tags()
		for _, t := range cTags {
			tagSet[t] = struct{}{}
		}
	}

	tags := make([]string, 0, len(tagSet))
	for t := range tagSet {
		tags = append(tags, t)
	}

	slices.Sort(tags)

	for i := range tags {
		tagMap[tags[i]] = i
	}

	return tags, tagMap
}

func intToStringOrEmpty(n int) string {
	if n == 0 {
		return ""
	}
	return strconv.Itoa(n)
}

func boolToStringOrEmpty(b bool) string {
	if b {
		return "Y"
	}
	return ""
}
