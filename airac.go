/*
 * Copyright (c) 2020 Johannes Kohnen <jwkohnen-github@ko-sys.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package airac

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	format        = "2006-01-02"
	cycleDuration = 24192e11 // 4 weeks
)

var (
	// nolint:gochecknoglobals
	_epoch = time.Date(1901, time.January, 10, 0, 0, 0, 0, time.UTC)
)

type (
	// AIRAC represents an Aeronautical Information Regulation And Control
	// (AIRAC) cycle.
	AIRAC uint16

	// Airac is a deprecated alias to AIRAC.
	Airac = AIRAC
)

// Effective returns the effective date of this AIRAC cycle.
func (a AIRAC) Effective() time.Time {
	return _epoch.Add(time.Duration(a) * cycleDuration)
}

// Year returns the year for this AIRAC cycle's identifier.
func (a AIRAC) Year() int {
	return a.Effective().Year()
}

// Ordinal returns the ordinal for this AIRAC cycle's identifier.
func (a AIRAC) Ordinal() int {
	return (a.Effective().YearDay()-1)/28 + 1
}

// FromDate returns the AIRAC cycle that occurred at date. A date before the
// internal epoch (1901-01-10) may return wrong data. The upper limit is year
// 2192.
func FromDate(date time.Time) AIRAC {
	a := date.Sub(_epoch) / cycleDuration
	return AIRAC(a)
}

// FromString returns an AIRAC cycle that matches the identifier <yyoo>, i.e.
// the last two digits of the year and the ordinal, each with leading zeros.
// This works for years between 1964 and 2063. Identifiers between "6401" and
// "9913" are interpreted as AIRAC cycles between the years 1964 and 1999
// inclusive. AIRAC cycles between "0001" and "6313" are interpreted as AIRAC
// cycles between the years 2000 and 2063 inclusive.
func FromString(yyoo string) (AIRAC, error) {
	year, ordinal, err := parseIdentifier(yyoo)
	if err != nil {
		return 0, err
	}

	lastAiracOfPreviousYear := FromDate(time.Date(year-1, time.December, 31, 0, 0, 0, 0, time.UTC))
	airac := lastAiracOfPreviousYear + AIRAC(ordinal)

	if airac.Year() != year {
		return 0, fmt.Errorf("illegal AIRAC id %q", yyoo)
	}

	return airac, nil
}

func parseIdentifier(yyoo string) (year, ordinal int, err error) {
	yyoo = strings.TrimSpace(yyoo)
	if len(yyoo) != 4 {
		return 0, 0, fmt.Errorf("illegal AIRAC id %q", yyoo)
	}

	if sign := yyoo[0]; sign == '+' || sign == '-' {
		return 0, 0, fmt.Errorf("illegal AIRAC id %q", yyoo)
	}

	yyooInt, err := strconv.Atoi(yyoo)
	if err != nil {
		return 0, 0, fmt.Errorf("illegal AIRAC id %q", yyoo)
	}

	year, ordinal = (yyooInt/100)+1900, yyooInt%100
	if year <= 1963 {
		year += 100
	}
	return year, ordinal, nil
}

// FromStringMust returns an AIRAC cycle that matches the identifier <yyoo>
// like FromString, but does not return an error. If there is an error it will
// panic instead.
func FromStringMust(yyoo string) AIRAC {
	airac, err := FromString(yyoo)
	if err != nil {
		panic(err)
	}
	return airac
}

// String returns a short representation of this AIRAC cycle. "YYOO"
func (a AIRAC) String() string {
	return fmt.Sprintf("%02d%02d", a.Year()%100, a.Ordinal())
}

// LongString returns a verbose representation of this AIRAC cycle.
// "YYOO (effective: YYYY-MM-DD; expires: YYYY-MM-DD)"
func (a AIRAC) LongString() string {
	n := a + 1
	return fmt.Sprintf("%02d%02d (effective: %s; expires: %s)",
		a.Year()%100,
		a.Ordinal(),
		a.Effective().Format(format),
		n.Effective().Add(-1).Format(format),
	)
}

// ByChrono is an []AIRAC wrapper, that satisfies sort.Interface and can be
// used to chronologically sort AIRAC instances.
type ByChrono []AIRAC

// Len ist the number of elements in the collection.
func (c ByChrono) Len() int { return len(c) }

// Less reports whether the element with index i should sort before the element
// with index j.
func (c ByChrono) Less(i, j int) bool { return c[i] < c[j] }

// Swap swaps the elements with indexes i and j.
func (c ByChrono) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

// static assert
var _ sort.Interface = (ByChrono)(nil)
