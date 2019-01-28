package main

import (
	"reflect"
	"testing"
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
