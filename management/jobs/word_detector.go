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
	//"unicode/utf8"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const MAX_WORD_LEN = 5
const ENTROPY_THRESHOLD = 2.5
const NUM_EVERY_TIME = 200000
const TIMES = 6

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

func DetectWord(items *[]Item) {
	//pattern := regexp.MustCompile(`[^-\p{Han}\w_\+]+`)
	pattern := regexp.MustCompile(`\P{Han}+`)
	total := 0
	suffixes := make([]string, 0)
	rSuffixes := make([]string, 0)
	fmt.Println("start")
	for _, v := range *items {
		sentences := pattern.Split(v.Title, -1)
		for _, s := range sentences {
			rs := []rune(strings.Trim(s, " "))
			lenrs := len(rs)
			if lenrs == 0 {
				continue
			}
			if lenrs < MAX_WORD_LEN+1 {
				suffixes = append(suffixes, s)
				rSuffixes = append(rSuffixes, ReverseString(s))
			} else {
				suffixes = append(suffixes, string(rs[0:MAX_WORD_LEN+1]))
				rSuffixes = append(rSuffixes, ReverseString((string(rs[lenrs-MAX_WORD_LEN-1 : lenrs]))))
			}
		}
	}
	fmt.Println("start sorting")
	sort.Sort(sort.StringSlice(suffixes))
	sort.Sort(sort.StringSlice(rSuffixes))
	fmt.Println("start counting freq and entropy")
	ls := len(suffixes)
	for l := 1; l <= MAX_WORD_LEN; l++ {
		pos := 0
		rs := []rune(suffixes[0])
		rwf := make(map[string]int)
		nTotal := 0
        pTotal := 0
		for pos < ls {
			crs := []rune(suffixes[pos])
			lenCrs := len(crs)
			if lenCrs >= l {
				if string(crs[0:l]) == string(rs[0:l]) {
					count := Freq[string(rs[0:l])]
					Freq[string(rs[0:l])] = count + 1
					total += 1
					if lenCrs > l {
						count = rwf[string(crs[l:l+1])]
						rwf[string(crs[l:l+1])] = count + 1
						pTotal++
					}
                    nTotal++
					pos++
					continue
				}
			}
			if len(rs) >= l {
				//if nTotal > 0 {
					Entropy[string(rs[0:l])] = CalEntropy(&rwf, nTotal, pTotal)
				//}
                /*else {
					Entropy[string(rs[0:l])] = ENTROPY_THRESHOLD
				}*/
			}
			rwf = make(map[string]int)
			nTotal = 0
            pTotal = 0
			if lenCrs < l {
				pos++
				if pos >= ls {
					break
				}
			}
			rs = []rune(suffixes[pos])
		}
		if len(rs) >= l {
			//if nTotal > 0 {
				Entropy[string(rs[0:l])] = CalEntropy(&rwf, nTotal, pTotal)
			//} else {
			//	Entropy[string(rs[0:l])] = 0
			//}
		}
	}

	ls = len(rSuffixes)
	for l := 1; l <= MAX_WORD_LEN; l++ {
		pos := 0
		rs := []rune(rSuffixes[0])
		rwf := make(map[string]int)
		nTotal := 0
		pTotal := 0
		for pos < ls {
			crs := []rune(rSuffixes[pos])
			lenCrs := len(crs)
			if lenCrs >= l {
				if string(crs[0:l]) == string(rs[0:l]) {
					if lenCrs > l {
						count := rwf[string(crs[l:l+1])]
						rwf[string(crs[l:l+1])] = count + 1
					    pTotal++
                    }
					nTotal++
					pos++
					continue
				}
			}
			if len(rs) > l {
				str := ReverseString(string(rs[0:l]))
				//if nTotal > 0 {
					Entropy[str] = math.Min(Entropy[str], CalEntropy(&rwf, nTotal, pTotal))
				//} else {
				//	Entropy[str] = math.Min(Entropy[str], ENTROPY_THRESHOLD)
				//}
			}
			rwf = make(map[string]int)
			nTotal = 0
			pTotal = 0
			if lenCrs < l {
				pos++
				if pos >= ls {
					break
				}
			}
			rs = []rune(rSuffixes[pos])

		}
		if len(rs) > l {
			str := ReverseString(string(rs[0:l]))
			//if nTotal > 0 {
				Entropy[str] = math.Min(Entropy[str], CalEntropy(&rwf, nTotal, pTotal))
			//} else {
			//	Entropy[str] = math.Min(Entropy[str], ENTROPY_THRESHOLD)
			//}
		}
	}

	for k, v := range Freq {
		Ps[k] = float64(v) / float64(total)
	}

	fmt.Println("start final word")
	session, err := mgo.Dial("10.0.1.23")
	if err != nil {
		panic(err)
	}
	wc := session.DB("words").C("dict")
	for k, v := range Ps {
		rk := []rune(k)
		lenrk := len(rk)
		if lenrk > 1 {
			p := float64(0)
			for i := 1; i < lenrk; i++ {
				t := Ps[string(rk[0:i])] * Ps[string(rk[i:])]
				p = math.Min(p, t)
			}
			/*err := fc.Insert(&word)
			  if err!=nil {
			      fmt.Println(err)
			  }*/
			//fmt.Println(Freq[k], v, p,  float64(v) / p,  Entropy[k])
			if Freq[k] >= 5 && float64(v)/p >= 100 && Entropy[k] >= ENTROPY_THRESHOLD {
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
}

func CalEntropy(wf *map[string]int, total int, hitTotal int) float64 {
	result := float64(0)
	for _, v := range *wf {
		p := float64(v) / float64(total)
		result -= p * math.Log(p)
	}
    notHitProb := float64(1) / float64(total)
    result -= notHitProb * math.Log(notHitProb) * float64(total - hitTotal)
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
		err = ic.Find(bson.M{"data_updated_time": bson.M{"$gt": startTime}, "extracted": nil}).Sort("data_updated_time").Limit(NUM_EVERY_TIME).Select(bson.M{"title": 1, "num_iid": 1, "data_updated_time": 1}).All(&items)
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
		/*err := ic.Update(bson.M{"num_iid" : v.NumIid}, bson.M{"$set" : bson.M{"extracted" : true}})
		  if err != nil {
		      fmt.Println(err)
		  }
		*/
	}
	fmt.Println("total", len(titles))
	if len(titles) > 900000 {
		DetectWord(&titles)
		session, err := mgo.Dial("10.0.1.23")
		if err != nil {
			panic(err)
		}
		pic := session.DB("words").C("process_info")
		pic.UpdateAll(bson.M{}, bson.M{"$set" : bson.M{"last_processed_timestamp": startTime}})
        /*
		for i := range titles {
			//fmt.Println(titles[i].NumIid)
			err := ic.Update(bson.M{"num_iid": titles[i].NumIid}, bson.M{"$set": bson.M{"extracted": true}})
			if err != nil {
				fmt.Println("Update err", err)
			}
		}*/
	}
}
