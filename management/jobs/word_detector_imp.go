package main

import (
	"fmt"
	//"io"
	"math"
	//"os"
	"regexp"
	"sort"
	"strings"
	"time"
    "unicode"
	"unicode/utf8"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const MAX_WORD_LEN = 5
const ENTROPY_THRESHOLD = 2.5
const NUM_EVERY_TIME = 50000
const TIMES = 12

var MgoDbName string = "mango"
var Freq = make(map[string]int)
var Ps = make(map[string]float64)
var Entropy = make(map[string]float64)

//var Words = make(map[string]bool)
//var FinalWords = make(map[string]bool)

type Item struct {
	NumIid          int       `bson:"num_iid"`
	Title           string    `bson:"title"`
	DataUpdatedTime time.Time `bson:"data_updated_time"`
}

func ReverseString(src string) string {
	rs := []rune(src)
	des := make([]rune, 0)
	for i := len(rs) - 1; i >= 0; i-- {
		des = append(des, rs[i])
	}
	return string(des)
}

type Word struct {
	Word string
	Freq int
	Prob float64
}

type Pos struct {
    Start int
    Len int
}

var AllStrings = make([]string, 0)
var AlphanumericMarks = make([]bool, 0)
var StringPos = make(map[string]Pos, 0)
var TotalLength int

type Suffixes []int

type ReverseSuffixes []int

func (v Suffixes) Len() int {
    return len(v)
}

func (v Suffixes) Swap(i, j int) {
    v[i], v[j] = v[j], v[i]
}

func (v Suffixes) Less(i, j int) bool {
    return CompareSuffixes(&v, i, j, MAX_WORD_LEN + 1) < 0
}

func (v ReverseSuffixes) Len() int {
    return len(v)
}

func (v ReverseSuffixes) Swap(i, j int) {
    v[i], v[j] = v[j], v[i]
}

func (v ReverseSuffixes) Less(i, j int) bool {
    return CompareReverseSuffixes(&v, i, j, MAX_WORD_LEN + 1) < 0
}

func CompareSuffixes(posArr *Suffixes, aPos int, bPos int, length int) int {
    for i := 0; i < length; i++ {
        if (*posArr)[aPos] + i >= TotalLength && (*posArr)[bPos] + i >= TotalLength {
            return 0
        }
        if (*posArr)[aPos] + i >= TotalLength {
            return -1
        }
        if (*posArr)[bPos] + i >= TotalLength {
            return 1
        }
        if AllStrings[(*posArr)[aPos] + i] <  AllStrings[(*posArr)[bPos] + i] {
            return -1
        }
        if AllStrings[(*posArr)[aPos] + i] >  AllStrings[(*posArr)[bPos] + i] {
            return 1
        }
    }
    return 0
}

func CompareReverseSuffixes(posArr *ReverseSuffixes, aPos int, bPos int, length int) int {
    for i := 0; i < length; i++ {
        if (*posArr)[aPos] - i < 0 && (*posArr)[bPos] - i < 0 {
            return 0
        }
        if (*posArr)[aPos] - i < 0 {
            return -1
        }
        if (*posArr)[bPos] - i < 0 {
            return 1
        }
        if AllStrings[(*posArr)[aPos] - i] <  AllStrings[(*posArr)[bPos] - i] {
            return -1
        }
        if AllStrings[(*posArr)[aPos] - i] >  AllStrings[(*posArr)[bPos] - i] {
            return 1
        }
    }
    return 0
}

func GetSuffixesString(pos int, length int, forward bool) string {
    if !forward {
        pos = pos - length + 1
    }
    res := ""
    lastAlphanumeric := false
    for i := 0; i < length && pos + i < TotalLength; i++ {
        if AlphanumericMarks[pos + i] {
            if lastAlphanumeric {
                res += " " + AllStrings[pos + i]
            } else {
                res += AllStrings[pos + i]
            }
            lastAlphanumeric = true
        } else {
            res += AllStrings[pos + i]
            lastAlphanumeric = false
        }
    }
    return res
}

func DetectWord(items *[]Item) {
	//pattern := regexp.MustCompile(`[^-\p{Han}\w_\+]+`)
	//pattern := regexp.MustCompile(`\P{Han}+`)
	pattern := regexp.MustCompile(`\s+`)
	suffixes := make(Suffixes, 0)
	rSuffixes := make(ReverseSuffixes, 0)
	fmt.Println("start")
    for _, v := range *items {
		sentences := pattern.Split(v.Title, -1)
		for _, s := range sentences {
			length := len(s)
            if length == 0 {
				continue
			}
            current := 0
            inAlphanumeric := true
            alphanumericStart := 0
            bs := []byte(s)
            for current < length {
                r, size := utf8.DecodeRune(bs[current:])
                cw := s[current : current + size]
                //fmt.Println("==", cw)
                if size <= 2 && (unicode.IsLetter(r) || unicode.IsNumber(r)) {
                    if !inAlphanumeric {
                        alphanumericStart = current
                        inAlphanumeric = true
                    }
                } else if !(cw == "-" || cw == "_" || cw == "+") {
                    if inAlphanumeric {
                        inAlphanumeric = false
                        if current != 0 {
                            AllStrings = append(AllStrings, strings.ToLower(s[alphanumericStart:current]))
                            AlphanumericMarks = append(AlphanumericMarks, true)
                        }
                    }
                    if !unicode.IsPunct(r) && !unicode.IsSymbol(r) {
                        AllStrings = append(AllStrings, s[current:current+size])
                        AlphanumericMarks = append(AlphanumericMarks, false)
                    }

                }
                current += size
            }
            if inAlphanumeric {
                if current != 0 {
                    AllStrings = append(AllStrings, strings.ToLower(s[alphanumericStart:current]))
                    AlphanumericMarks = append(AlphanumericMarks, true)
                    TotalLength++
                }
            }
		}
	}
    fmt.Println(TotalLength)
	TotalLength = len(AllStrings)
    for i := 0; i < TotalLength - MAX_WORD_LEN; i++ {
        suffixes = append(suffixes, i)
        rSuffixes = append(rSuffixes, i + MAX_WORD_LEN)
    }
    sort.Sort(Suffixes(suffixes))
	sort.Sort(ReverseSuffixes(rSuffixes))
	fmt.Println("start counting freq and entropy")
	ls := len(suffixes)
    /*for i := 0; i < ls; i++ {
        fmt.Println(GetSuffixesString(suffixes[i], MAX_WORD_LEN, true))
    }
    for i := 0; i < ls; i++ {
        fmt.Println(GetSuffixesString(rSuffixes[i], MAX_WORD_LEN, false))
    }*/
    total := 0
	for l := 1; l <= MAX_WORD_LEN; l++ {
        fmt.Println("current len", l)
		pos := 0
		lastPos := 0
        lastString := GetSuffixesString(suffixes[lastPos], l, true)
        StringPos[lastString] = Pos{Start: suffixes[lastPos], Len : l}
        rwf := make(map[string]int)
		nTotal := 0
		for pos < ls {
            if CompareSuffixes(&suffixes, lastPos, pos, l) == 0 {
                count := Freq[lastString]
                Freq[lastString] = count + 1
                total++
                count = rwf[AllStrings[suffixes[pos] + l]]
                rwf[AllStrings[suffixes[pos] + l]] = count + 1
                nTotal++
                pos++
                continue
            }
			Entropy[lastString] = CalEntropy(&rwf, nTotal)
			rwf = make(map[string]int)
			nTotal = 0
			lastPos = pos
            lastString = GetSuffixesString(suffixes[lastPos], l, true)
            StringPos[lastString] = Pos{Start: suffixes[lastPos], Len : l}
		}
		Entropy[lastString] = CalEntropy(&rwf, nTotal)
	}
    fmt.Println("1111111111111111111111111111111111111")
	ls = len(rSuffixes)
	for l := 1; l <= MAX_WORD_LEN; l++ {
        fmt.Println("current len", l)
		pos := 0
		lastPos := 0
        lastString := GetSuffixesString(rSuffixes[lastPos], l, false)
        rwf := make(map[string]int)
		nTotal := 0
		for pos < ls {
            if CompareReverseSuffixes(&rSuffixes, lastPos, pos, l) == 0 {
                count := rwf[AllStrings[rSuffixes[pos] - l]]
                rwf[AllStrings[rSuffixes[pos] - l]] = count + 1
                nTotal++
                pos++
                continue
            }
			Entropy[lastString] = math.Min(Entropy[lastString], CalEntropy(&rwf, nTotal))
			rwf = make(map[string]int)
			nTotal = 0
			lastPos = pos
            lastString = GetSuffixesString(rSuffixes[lastPos], l, false)
		}
		Entropy[lastString] = math.Min(Entropy[lastString], CalEntropy(&rwf, nTotal))
	}
    fmt.Println("1111111111111111111111111111111111111")

	for k, v := range Freq {
		Ps[k] = float64(v) / float64(total)
	}

	fmt.Println("start final word")
	session, err := mgo.Dial("10.0.1.23")
	if err != nil {
		panic(err)
	}
	wc := session.DB("words").C("dict_chi_eng")
	for k, v := range Ps {
		stringPos := StringPos[k]
        rk := []rune(k)
        if stringPos.Len > 1 || len(rk) > 1  {
			p := float64(0)
            if stringPos.Len > 1 {
			    for i := 1; i < stringPos.Len; i++ {
				    t := Ps[GetSuffixesString(stringPos.Start, i, true)]  * Ps[GetSuffixesString(stringPos.Start + i, stringPos.Len - i, true)]
				    p = math.Min(p, t)
			    }
            }
			//fmt.Println(Freq[k], v, p,  float64(v) / p,  Entropy[k])
			if ((stringPos.Len > 1 && float64(v)/p >= 120 && Freq[k] >= 10) || (stringPos.Len == 1 && len(rk) > 1 && Freq[k] >= 100)) && Entropy[k] >= ENTROPY_THRESHOLD {
				oldWord := Word{}
				err := wc.Find(bson.M{"word": k}).One(&oldWord)
				if err != nil {
					if err.Error() == "not found" {
						word := Word{Word: k, Freq: Freq[k], Prob: v}
						wc.Insert(&word)
					} else {
						fmt.Println("find word error", err)
						fmt.Println("miss word", k, Freq[k], v, p, Entropy[k])
					}
				} else {
					freq := oldWord.Freq + Freq[k]
					prob := float64(oldWord.Freq+Freq[k]) / (float64(oldWord.Freq)/oldWord.Prob + float64(total))
					wc.Update(bson.M{"word": k}, bson.M{"$set": bson.M{"freq": freq, "prob": prob}})
				}
			}
		}
	}
    session.Close()
}

func CalEntropy(wf *map[string]int, total int) float64 {
	result := float64(0)
	for _, v := range *wf {
		p := float64(v) / float64(total)
		result -= p * math.Log(p)
	}
	return result
}

func main() {

    s, err := mgo.Dial("10.0.1.23")
    if err != nil {
        panic(err)
    }
    pic := s.DB("words").C("process_info")
    res := bson.M{}
    pic.Find(nil).One(&res)
	//startTime := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	startTime := res["last_processed_timestamp"].(time.Time)
	titles := make([]Item, 0)
	total := 0
	for total < TIMES {
		session, err := mgo.Dial("10.0.1.23")
		if err != nil {
			panic(err)
		}
		ic := session.DB(MgoDbName).C("taobao_items_depot")
		items := make([]Item, 0)
		fmt.Println(startTime)
		err = ic.Find(bson.M{"data_updated_time": bson.M{"$gt": startTime}}).Sort("data_updated_time").Limit(NUM_EVERY_TIME).Select(bson.M{"title": 1, "num_iid": 1, "data_updated_time": 1}).All(&items)
		if err != nil {
			fmt.Println(err)
		}
		l := len(items)
		if l == 0 {
			break
		}
		fmt.Println("items", l)
		total++
		titles = append(titles, items...)
		startTime = items[l-1].DataUpdatedTime
		fmt.Println(startTime)
        session.Close()
	}
	fmt.Println("total", len(titles))
	if len(titles) > 500000 {
		DetectWord(&titles)
		session, err := mgo.Dial("10.0.1.23")
		if err != nil {
			panic(err)
		}
		pic := session.DB("words").C("process_info")
		pic.UpdateAll(bson.M{}, bson.M{"$set" : bson.M{"last_processed_timestamp": startTime}})
        session.Close()
	}
}
