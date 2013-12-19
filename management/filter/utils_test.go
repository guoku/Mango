package filter

import (
	"log"
	"testing"
)

func TestSliceToString(t *testing.T) {
	s := SplitTextToWords([]byte("我在中国China北京12389 America美国"))
	log.Println(TextSliceToString(s))
}
