package main
import (
    "fmt"
    //"io"
    "math"
    //"os"
    "regexp"
    "strings"
    "sort"
    "time"
    //"unicode/utf8"
    "labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const MAX_WORD_LEN = 5
const ENTROPY_THRESHOLD = 2.5

var MgoSession *mgo.Session
var MgoDbName string = "mango"
var Freq = make(map[string]int)
var Ps = make(map[string]float64)
var Entropy = make(map[string]float64)
//var Words = make(map[string]bool)
//var FinalWords = make(map[string]bool)
type Item struct {
    Title string
}

func init() {
    session, err := mgo.Dial("10.0.1.23")
    if err != nil {
        panic(err)
    }
    MgoSession = session
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
    Pmi float64
    Entropy float64
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
            if lenrs < MAX_WORD_LEN + 1 {
                suffixes = append(suffixes, s)
                rSuffixes = append(rSuffixes, ReverseString(s))
            } else {
                suffixes = append(suffixes, string(rs[0 : MAX_WORD_LEN + 1]))
                rSuffixes = append(rSuffixes, ReverseString((string(rs[lenrs - MAX_WORD_LEN - 1 : lenrs]))))
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
        for pos < ls {
            crs := []rune(suffixes[pos])
            lenCrs := len(crs)
            if lenCrs >= l {
                if string(crs[0 : l]) == string(rs[0 : l]) {
                    count := Freq[string(rs[0 : l])]
                    Freq[string(rs[0 : l])] = count + 1
                    total += 1
                    if lenCrs > l {
                        count = rwf[string(crs[l : l + 1])]
                        rwf[string(crs[l : l + 1])] = count + 1
                        nTotal ++
                    }
                    pos++
                    continue
                }
            }
            if len(rs) >= l {
                if nTotal > 0 {
                    Entropy[string(rs[0 : l])] = CalEntropy(&rwf, nTotal)
                } else {
                    Entropy[string(rs[0 : l])] = ENTROPY_THRESHOLD
                }
            }
            rwf = make(map[string]int)
            nTotal = 0
            if lenCrs < l {
                pos ++
                if pos >= ls {
                    break
                }
            }
            rs = []rune(suffixes[pos])
        }
        if len(rs) >= l {
            if nTotal > 0 {
                Entropy[string(rs[0 : l])] = CalEntropy(&rwf, nTotal)
            } else {
                Entropy[string(rs[0 : l])] = 0
            }
        }
    }

    ls = len(rSuffixes)
    for l := 1; l <= MAX_WORD_LEN; l++ {
        pos := 0
        rs := []rune(rSuffixes[0])
        rwf := make(map[string]int)
        nTotal := 0
        for pos < ls {
            crs := []rune(rSuffixes[pos])
            lenCrs := len(crs)
            if lenCrs >= l {
                if string(crs[0 : l]) == string(rs[0 : l]) {
                    if lenCrs > l {
                        count := rwf[string(crs[l : l + 1])]
                        rwf[string(crs[l : l + 1])] = count + 1
                        nTotal ++
                    }
                    pos++
                    continue
                }
            }
            if len(rs) > l {
                str := ReverseString(string(rs[0 : l]))
                if nTotal > 0 {
                    Entropy[str] = math.Min(Entropy[str], CalEntropy(&rwf, nTotal))
                } else {
                    Entropy[str] = math.Min(Entropy[str], ENTROPY_THRESHOLD)
                }
            }
            rwf = make(map[string]int)
            nTotal = 0
            if lenCrs < l {
                pos ++
                if pos >= ls {
                    break
                }
            }
            rs = []rune(rSuffixes[pos])

        }
        if len(rs) > l {
            str := ReverseString(string(rs[0 : l]))
            if nTotal > 0 {
                Entropy[str] = math.Min(Entropy[str], CalEntropy(&rwf, nTotal))
            } else {
                Entropy[str] = math.Min(Entropy[str], ENTROPY_THRESHOLD)
            }
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
    fc := session.DB("words").C("frequency")
    wc := session.DB("words").C("dict")
    for k, v := range Ps {
        rk := []rune(k)
        lenrk := len(rk)
        if lenrk > 1 {
            p := float64(0)
            for i := 1; i < lenrk; i++ {
                t := Ps[string(rk[0 : i ])] * Ps[string(rk[i:])]
                p = math.Min(p, t)
            }
            word := Word{Word : k, Freq: Freq[k], Prob : v, Pmi : p, Entropy : Entropy[k]}
            err := fc.Insert(&word)
            if err!=nil {
                fmt.Println(err)
            }
            //fmt.Println(Freq[k], v, p,  float64(v) / p,  Entropy[k])
            if Freq[k] >= 5 && float64(v) / p > 80 && Entropy[k] > ENTROPY_THRESHOLD {
                wc.Insert(&word)
            }
        }
    }
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
    items := make([]Item, 0)
    ic := MgoSession.DB(MgoDbName).C("taobao_items_depot")
    startTime :=  time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
    ic.Find(bson.M{"data_updated_time" : bson.M{"$gt" : startTime}}).Select(bson.M{"title": 1}).All(&items)
    DetectWord(&items)
}

