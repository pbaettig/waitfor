package main

import (
	"reflect"
	"regexp"
	"sort"
	"testing"

	"github.com/pbaettig/waitfor/internal/app"
)

func TestHttpCodeListParse(t *testing.T) {
	table := []struct {
		in   string
		want []int
		err  bool
	}{
		{"200,201,400-404", []int{200, 201, 400, 401, 402, 403, 404}, false},
		{"503", []int{503}, false},
		{"401, 404-404", []int{401, 404}, false},
		{"abcd,def,200a-300", []int{}, true},
		{"300-200", []int{}, true},
		{"", []int{}, true},
	}

	for _, testCase := range table {
		got, err := parseCodes(testCase.in)

		if err == nil && testCase.err {
			// We don't have an error but should have one
			t.Logf("Test for httpCodeList.Parse(\"%s\") failed, wanted an error but got nil", testCase.in)
			t.FailNow()
		}
		if err != nil && !testCase.err {
			t.Logf(err.Error())
			// We  have an error but shouldn't have one
			t.Logf("Test for httpCodeList.Parse(\"%s\") failed, wanted no error but got %s", testCase.in, err)
			t.FailNow()
		}
		if len(testCase.want) > 0 {
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("Test for httpCodeList.Parse(\"%s\") failed, wanted %v but got %v", testCase.in, testCase.want, got)
			}
		}

	}
}

func compareIntSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Ints(a)
	sort.Ints(b)
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestHttpSpecParse(t *testing.T) {
	table := []struct {
		in   string
		want app.HTTPWait
		err  bool
	}{
		{
			"http://localhost:8080/1234|400-404|Hi .*",
			app.HTTPWait{
				URL:                   "http://localhost:8080/1234",
				AcceptableStatusCodes: []int{400, 401, 402, 403, 404},
				ContentMatch:          regexp.MustCompile(`Hi .*`),
			},
			false,
		},
		{
			"http://localhost:8080/1234||",
			app.HTTPWait{
				URL:                   "http://localhost:8080/1234",
				AcceptableStatusCodes: []int{},
				ContentMatch:          nil,
			},
			true,
		},
		{
			"http://localhost:8080/1234|400-200",
			app.HTTPWait{
				URL:                   "http://localhost:8080/1234",
				AcceptableStatusCodes: []int{},
				ContentMatch:          nil,
			},
			true,
		},
		{
			"localhost:8080/1234|200-20a",
			app.HTTPWait{
				URL:                   "http://localhost:8080/1234",
				AcceptableStatusCodes: []int{},
				ContentMatch:          nil,
			},
			true,
		},
		{
			"localhost:8080/abcde/fgh",
			app.HTTPWait{
				URL:                   "http://localhost:8080/abcde/fgh",
				AcceptableStatusCodes: []int{},
				ContentMatch:          nil,
			},
			false,
		},
		{
			"https://localhost/abcde/fgh|203",
			app.HTTPWait{
				URL:                   "https://localhost/abcde/fgh",
				AcceptableStatusCodes: []int{203},
				ContentMatch:          nil,
			},
			false,
		},
		{
			"https://localhost/abcde/fgh|203,204-207,404|Welcome to .*",
			app.HTTPWait{
				URL:                   "https://localhost/abcde/fgh",
				AcceptableStatusCodes: []int{203, 204, 205, 206, 207, 404},
				ContentMatch:          regexp.MustCompile(`Welcome to .*`),
			},
			false,
		},
	}

	for _, test := range table {
		htws := make(httpWaits, 0)
		err := htws.Set(test.in)
		if err == nil && test.err {
			// have no error but want one
			t.Logf("Test failed, wanted error but got none")
			t.FailNow()
		} else if err != nil && !test.err {
			// have an error but don't want one
			t.Logf("Test failed, got error '%s' but didn't want one", err)
			t.FailNow()
		}

		if !test.err {
			got := htws[0]
			if got.URL != test.want.URL {
				t.Errorf("Test failed, wanted URL: %s but got %s", test.want.URL, got.URL)
			}

			if !compareIntSlice(test.want.AcceptableStatusCodes, got.AcceptableStatusCodes) {
				t.Errorf("Test failed, wanted AcceptableStatusCodes: %v, got %v", test.want.AcceptableStatusCodes, got.AcceptableStatusCodes)
			}

			if !reflect.DeepEqual(test.want.ContentMatch, got.ContentMatch) {
				t.Errorf("Test failed, wanted ContentMatch: %s, got %s", test.want.ContentMatch, got.ContentMatch)
			}

		}
	}

}
