package main
import (
    "fmt"
    "math"
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
var Words = make(map[string]bool)
var FinalWords = make(map[string]bool)
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
func DetectWord(items *[]Item) {
    //pattern := regexp.MustCompile(`[^-\p{Han}\w_\+]+`)
    pattern := regexp.MustCompile(`\P{Han}+`)
    //total := 0
    //allSentences := make([]string, 0)
    suffixes := make([]string, 0)
    rSuffixes := make([]string, 0)
    for _, v := range *items {
        sentences := pattern.Split(v.Title, -1)
        for _, s := range sentences {
            rs := []rune(strings.Trim(s, " "))
            lenrs := len(rs)
            if lenrs == 0 {
                continue
            }
            /*
            for i := 1; i <= MAX_WORD_LEN; i++ {
                for j := 0; j <= lenrs - i; j++ {
                    count := Freq[string(rs[j : j + i])]
                    Freq[string(rs[j : j + i])] = count + 1
                    total += 1
                }
            }
            */
            if lenrs < MAX_WORD_LEN + 1 {
                suffixes = append(suffixes, s)
                rSuffixes = append(rSuffixes, ReverseString(s))
            } else {
                suffixes = append(suffixes, string(rs[0 : MAX_WORD_LEN + 1]))
                suffixes = append(suffixes, ReverseString((string(rs[lenrs - MAX_WORD_LEN - 1 : lenrs]))))
            }
        }
        //allSentences = append(allSentences, sentences...)
    }
    sort.Sort(sort.StringSlice(suffixes))
    for _, v := range suffixes {
        fmt.Println(v)
    }
    fmt.Println("111111111111111111111111111111111111")
    /*
    for k, v := range Freq {
        Ps[k] = float64(v) / float64(total)
    }

    for k, v := range Ps {
        rk := []rune(k)
        lenrk := len(rk)
        if lenrk > 1 {
            p := float64(0)
            for i := 0; i < lenrk; i++ {
                t := Ps[string(rk[0 : i ])] * Ps[string(rk[i:])]
                p = math.Max(p, t)
            }
            if Freq[k] >= 3 && float64(v) / p > 100 {
                Words[k] = true
            }
        }
    }
    fmt.Println("1222222222222222222222222222222221")
    lwf := make(map[string]int)
    rwf := make(map[string]int)
    for k := range Words {
        lf := true
        rf := true
        rk := []rune(k)
        ltotal := 0
        rtotal := 0
        spat := regexp.MustCompile(fmt.Sprintf(".?%s.?", k))
        for _, st := range allSentences {
            matches := spat.FindAllString(st, -1)
            for i := range matches {
                rm := []rune(matches[i])
                if rm[0] != rk[0] {
                    cnt := lwf[string(rm[0])]
                    lwf[string(rm[0])] = cnt + 1
                    ltotal++
                } else {
                    lf = false
                }
                if rm[len(rm) - 1] != rk[len(rk) - 1] {
                    cnt := rwf[string(rm[len(rm) - 1])]
                    rwf[string(rm[len(rm) - 1])] = cnt + 1
                    rtotal++
                } else {
                    rf = false
                }
            }
        }
        leftEntropy := CalEntropy(&lwf, ltotal)
        rightEntropy := CalEntropy(&rwf, rtotal)
        if lf && len(lwf) >0 && leftEntropy > ENTROPY_THRESHOLD {
            continue
        }
        if rf && len(rwf) >0 && rightEntropy > ENTROPY_THRESHOLD {
            continue
        }
        FinalWords[k] = true
    }

    tt := 0
    for k := range FinalWords {
        tt ++
        fmt.Println(k)
    }
    fmt.Println(tt)
    */
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
    ic.Find(bson.M{"data_updated_time" : bson.M{"$gt" : startTime}}).Select(bson.M{"title": 1}).Limit(500).All(&items)
    DetectWord(&items)
}

