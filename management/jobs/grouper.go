package main


import (
    "fmt"
    "math"
    "time"

    "Mango/management/filter"
    "Mango/management/models"
    "Mango/management/segment"
    "github.com/qiniu/log"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
)

var MgoSession *mgo.Session
var NumEveryTime int = 10000
var segmenter *segment.GuokuSegmenter
var totalWords = 100000000
var totalDocs = 5000000
var threshold float64 = 0.8
var filterTree *filter.TrieTree

func init() {
    sess, err := mgo.Dial("10.0.1.23")
    if err != nil {
        log.Fatal("Can not reach mongodb")
    }
    MgoSession = sess
    segmenter = &segment.GuokuSegmenter{}
    segmenter.LoadDictionary()
    filterTree = &filter.TrieTree{}
    filterTree.LoadDictionary("10.0.1.23", "words", "brands")
    filterTree.LoadBlackWords("10.0.1.23", "words", "dict_chi_eng")
}

func getStartTime() time.Time {
    c := MgoSession.DB("mango").C("process_info")
    result := bson.M{}
    c.Find(bson.M{"name" : "grouper"}).One(&result)
    startTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
    if _, present := result["start_time"]; present {
        startTime = result["start_time"].(time.Time)
    }
    return startTime
}

func saveStartTime(startTime time.Time) {
    c := MgoSession.DB("mango").C("process_info")
    c.Upsert(bson.M{"name" : "grouper"}, bson.M{"$set" : bson.M{"start_time" : startTime}})
}

func ExtractItemVector(item models.TaobaoItemStd) map[string]float64 {
    segs, freqs:= segmenter.GetSegmentAndFrequency(item.Title)
    segStrings := filter.SegSliceToSegString(segs)
    filteredSegs := filterTree.FiltrateForArray(segs)
    filteredSegStrings := filter.SegSliceToSegString(filteredSegs)
    cnt := 0
    // delete the freq of filtered segs
    fmt.Println(segStrings)
    fmt.Println(freqs)
    fmt.Println(filteredSegStrings)

    newFreqs := make([]int, 0)
    flen := len(filteredSegStrings)
    for i := range segStrings {
        if cnt == flen {
            break
        }
        if segStrings[i] == filteredSegStrings[cnt] {
            cnt ++
            newFreqs = append(newFreqs, freqs[i])
        }
    }
    fmt.Println(newFreqs)
    m := make(map[string]float64)
    for i := range filteredSegStrings {
        m[filteredSegStrings[i]] = m[filteredSegStrings[i]] + getWeight(newFreqs[i])
    }

    if item.Props["品牌"] != "" {
        bsegs := filter.SegSliceToSegString(segmenter.Segment(item.Props["品牌"]))
        for _, v := range bsegs {
            m[v] = 0.00005
        }
    }
    log.Println(m)
    /*
    vector := make([]models.VectorItem, len(m))

    for k, v := range m {
        vector[i] = models.VectorItem{Name : k, Weight: v}
    }
    return vector*/
    return m
}

func getWeight(freq int) float64{
    // approximate tf-idf value
    tf := float64(freq) / float64(totalWords)
    idf := math.Log(float64(totalDocs) / float64(freq+1))
    return tf * idf
}

func CompareVector(vector1, vector2 map[string]float64) float64 {
    // get cosine similiarity of two vectors
    allWords := make(map[string]bool)
    for k := range vector1 {
        allWords[k] = true
    }
    for k := range vector2 {
        allWords[k] = true
    }
    var s float64
    var s1 float64
    var s2 float64
    for k := range allWords {
        s += vector1[k] * vector2[k]
        s1 += vector1[k] * vector1[k]
        s2 += vector2[k] * vector2[k]
    }
    return s / (math.Sqrt(s1) * math.Sqrt(s2))
}

func priceDiff(p1, p2 float32) float32 {
    if p1 < p2 {
        p1, p2 = p2, p1
    }
    return 1 - p2 / p1
}


func generateNewGroup(item models.TaobaoItemStd, vector map[string]float64) {
    gtc := MgoSession.DB("mango").C("metadata")
    gc := MgoSession.DB("mango").C("item_group")
    ic := MgoSession.DB("mango").C("taobao_items_depot")
    result := bson.M{}
    gtc.Find(bson.M{"type" : "group_info"}).One(&result)
    var lastGroupId int
    if _, present := result["last_group_id"]; !present {
        lastGroupId = 0
    } else {
        lastGroupId = result["last_group_id"].(int)
    }
    vectorFreq := make(map[string]int)
    for k := range vector {
        vectorFreq[k] = 1
    }
    g := models.ItemGroup{GroupId : lastGroupId + 1, Vector : vector, VectorFreq : vectorFreq,
Status : "new", TaobaoCid : item.Cid, NumItem : 1, AveragePrice : item.Price, DelegateId : item.NumIid }
    log.Println("new group",  g)
    gtc.Upsert(bson.M{"type" : "group_info"}, bson.M{"$set": bson.M{"last_group_id" : lastGroupId + 1}})
    err := gc.Insert(&g)
    log.Println("____________________", err)
    ic.Update(bson.M{"num_iid" : item.NumIid}, bson.M{"$set" : bson.M{"group_id" : g.GroupId}})
}

func scanTaobaoItems(startTime time.Time) {
    ic := MgoSession.DB("mango").C("taobao_items_depot")
    gc := MgoSession.DB("mango").C("item_group")
    items := make([]models.TaobaoItemStd, 0)
    ic.Find(bson.M{"data_updated_time" : bson.M{"$gt" : startTime}}).Sort("data_updated_time", "cid").Limit(NumEveryTime).All(&items)
    currentCid := -1
    groups := make([]models.ItemGroup, 0)
    log.Println("start process")
    for i := range items {
        // extract vector
        // find matched cid group
        // loop
        //      compare delegate item and group vector
        //      if similar(delegate, item) > xxxx && similar(group_vector, item) > xxxx, max_group = group
        //          if similar(group_vector, item) > similar(group_vector, delegate) delegate = item
        //  if max > xxx add to it 
        //  else create a new group
        if currentCid != items[i].Cid {
            currentCid = items[i].Cid
            groups = make([]models.ItemGroup, 0)
            gc.Find(bson.M{"taobao_cid" : currentCid}).All(&groups)
        }
        maxSimilarity := float64(0)
        var maxGroup *models.ItemGroup
        itemVector := ExtractItemVector(items[i])
        for i, group := range groups {
            if priceDiff(group.AveragePrice, items[i].Price) > 0.35 {
                continue
            }
            delegateItem := models.TaobaoItemStd{}
            err := ic.Find(bson.M{"num_iid": group.DelegateId}).One(&delegateItem)
            if err != nil {
                log.Error("Can't find delegate item:", group.DelegateId)
                continue
            }
            delegateVector := ExtractItemVector(delegateItem)
            sim := CompareVector(itemVector, delegateVector)
            log.Println("sim", sim)
            if sim > maxSimilarity {
                maxSimilarity = sim
                maxGroup = &groups[i]
            }
            aSim := CompareVector(group.Vector, itemVector)
            log.Println("asim", aSim)
            if aSim > maxSimilarity {
                maxSimilarity = aSim
                maxGroup = &groups[i]
            }
            log.Println("max", maxSimilarity)
        }
        log.Println("===================", maxSimilarity)
        if maxSimilarity < threshold {
            log.Println("not found")
            generateNewGroup(items[i], itemVector)
        } else {
            log.Println("found", maxGroup.GroupId)
            ic.Update(bson.M{"num_iid": items[i].NumIid}, bson.M{"$set" : bson.M{"group_id": maxGroup.GroupId}})
            //maxG
        }
    }
}

func main() {
    defer MgoSession.Close()
    for {
        startTime := getStartTime()
        log.Println("start", startTime)
        scanTaobaoItems(startTime)
        saveStartTime(startTime)
        time.Sleep(1 * time.Minute)
    }
}

