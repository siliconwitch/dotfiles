package footage

import (
	"reflect"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name  string
		words []Word
		want  []Phrase
	}{
		{"one phrase", []Word{{"so", 0, 0.2, false}, {"today", 0.3, 0.6, false}}, []Phrase{{0, 1}}},
		{"a pause of 0.3 s", []Word{{"so", 0, 0.2, false}, {"today", 0.5, 0.8, false}}, []Phrase{{0, 0}, {1, 1}}},
		{"a sentence end", []Word{{"done.", 0, 0.2, false}, {"next", 0.25, 0.5, false}}, []Phrase{{0, 0}, {1, 1}}},
		{"a sound", []Word{{"so", 0, 0.2, false}, {"(laughs)", 0.2, 0.9, true}, {"yes", 0.9, 1.1, false}}, []Phrase{{0, 0}, {1, 1}, {2, 2}}},
		{"a long word", []Word{{"so", 0, 0.2, false}, {"blah", 0.2, 5.5, false}, {"yes", 5.5, 5.7, false}}, []Phrase{{0, 0}, {1, 1}, {2, 2}}},
		{"no words", nil, nil},
	}

	for _, test := range tests {
		got := Split(test.words)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s: got %v, want %v", test.name, got, test.want)
		}
	}
}

func TestLines(t *testing.T) {
	words := []Word{{"Hello.", 1, 1.4, false}, {"Blah.", 4, 6, false}}
	want := "p1 1.00-1.40 Hello.\n-- 2.6 s without speech --\np2 4.00-6.00 Blah. (\"Blah.\" lasts 2.0 s)\n"

	got := Lines(words, Split(words))
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
