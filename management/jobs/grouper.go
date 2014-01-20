package main

import (
	//"fmt"
	"math"
	"sort"
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
var threshold float64 = 0.995
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
	c.Find(bson.M{"name": "grouper"}).One(&result)
	startTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, present := result["start_time"]; present {
		startTime = result["start_time"].(time.Time)
	}
	return startTime
}

func saveStartTime(startTime time.Time) {
	c := MgoSession.DB("mango").C("process_info")
	c.Upsert(bson.M{"name": "grouper"}, bson.M{"$set": bson.M{"start_time": startTime}})
}

func ExtractItemVector(item models.TaobaoItemStd) map[string]float64 {
	segs, freqs := segmenter.GetSegmentAndFrequency(item.Title)
	segStrings := filter.SegSliceToSegString(segs)
	filteredSegs := filterTree.FiltrateForArray(segs)
	filteredSegStrings := filter.SegSliceToSegString(filteredSegs)
	cnt := 0
	// delete the freq of filtered segs

	newFreqs := make([]int, 0)
	flen := len(filteredSegStrings)
	for i := range segStrings {
		if cnt == flen {
			break
		}
		if segStrings[i] == filteredSegStrings[cnt] {
			cnt++
			newFreqs = append(newFreqs, freqs[i])
		}
	}
	m := make(map[string]float64)
	for i := range filteredSegStrings {
        r := []rune(filteredSegStrings[i])
        if len(r) < 2 || newFreqs[i] < 10 {
            continue
        }
		m[filteredSegStrings[i]] = m[filteredSegStrings[i]] + getWeight(newFreqs[i])
	}

	if item.Props["品牌"] != "" {
		bsegs := filter.SegSliceToSegString(segmenter.Segment(item.Props["品牌"]))
		for _, v := range bsegs {
			m[v] = 0.00005
		}
	}
	/*
	   vector := make([]models.VectorItem, len(m))

	   for k, v := range m {
	       vector[i] = models.VectorItem{Name : k, Weight: v}
	   }
	   return vector*/
	return m
}

func getWeight(freq int) float64 {
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
	return 1 - p2/p1
}

func generateNewGroup(item models.TaobaoItemStd, vector map[string]float64) models.ItemGroup {
	gtc := MgoSession.DB("mango").C("metadata")
	gc := MgoSession.DB("mango").C("item_group")
	ic := MgoSession.DB("mango").C("taobao_items_depot")
	result := bson.M{}
	gtc.Find(bson.M{"type": "group_info"}).One(&result)
	var lastGroupId int
	if _, present := result["last_group_id"]; !present {
		lastGroupId = 0
	} else {
		lastGroupId = result["last_group_id"].(int)
	}
	g := models.ItemGroup{GroupId: lastGroupId + 1, Vector: vector,
		Status: "new", TaobaoCid: item.Cid, NumItem: 1, AveragePrice: item.Price, DelegateId: item.NumIid}
	gtc.Upsert(bson.M{"type": "group_info"}, bson.M{"$set": bson.M{"last_group_id": lastGroupId + 1}})
	err := gc.Insert(&g)
	if err != nil {
		log.Println("insert new group error", err)
	}
	ic.Update(bson.M{"num_iid": item.NumIid}, bson.M{"$set": bson.M{"group_id": g.GroupId}})
    return g
}

type mapItem struct {
	Key   string
	Value float64
}

type mapItemSlice []mapItem

func (this mapItemSlice) Len() int {
	return len(this)
}

func (this mapItemSlice) Less(i, j int) bool {
	return this[i].Value > this[j].Value
}

func (this mapItemSlice) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func mergeVector(averageVector map[string]float64, numItem int, newVector map[string]float64) {
	for k := range newVector {
		averageVector[k] = (averageVector[k]*float64(numItem) + newVector[k]) / float64(numItem+1)
	}
	// Only preserve the 20 keys with the biggest values
	if len(averageVector) > 20 {
		items := make([]mapItem, 0)
		for k, v := range averageVector {
			items = append(items, mapItem{Key: k, Value: v})
		}
		sort.Sort(mapItemSlice(items))
		for i := range items {
			if i < 20 {
				continue
			}
			delete(averageVector, items[i].Key)
		}
	}
}

type TaobaoItemSlice []models.TaobaoItemStd

func (this TaobaoItemSlice) Len() int {
    return len(this)
}

func (this TaobaoItemSlice) Less(i, j int) bool {
    return this[i].Cid < this[j].Cid
}

func (this TaobaoItemSlice) Swap(i, j int) {
    this[i], this[j] = this[j], this[i]
}

func scanTaobaoItems(startTime time.Time) time.Time{
	ic := MgoSession.DB("mango").C("taobao_items_depot")
	gc := MgoSession.DB("mango").C("item_group")
	items := make([]models.TaobaoItemStd, 0)
	ic.Find(bson.M{"data_updated_time": bson.M{"$gt": startTime}}).Sort("data_updated_time").Limit(NumEveryTime).All(&items)
	currentCid := -1
	groups := make([]models.ItemGroup, 0)
	log.Println("start process")
    lastStartTime := items[len(items) - 1].DataUpdatedTime
    sort.Sort(TaobaoItemSlice(items))

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
			gc.Find(bson.M{"taobao_cid": currentCid}).All(&groups)
		}
		maxSimilarity := float64(0)
		var maxGroup *models.ItemGroup
		itemVector := ExtractItemVector(items[i])
		for j, group := range groups {
			if priceDiff(group.AveragePrice, items[i].Price) > 0.3 {
                continue
			}
			delegateItem := models.TaobaoItemStd{}
			err := ic.Find(bson.M{"num_iid": group.DelegateId}).One(&delegateItem)
			if err != nil {
				log.Error("Can't find delegate item:", group.DelegateId)
				continue
			}
            /*
			delegateVector := groups[j].DelegateVector
			sim := CompareVector(itemVector, delegateVector)
			if sim > maxSimilarity {
				maxSimilarity = sim
				maxGroup = &groups[j]
			}
            */
			aSim := CompareVector(group.Vector, itemVector)
			if aSim > maxSimilarity {
				maxSimilarity = aSim
				maxGroup = &groups[j]
			}
		}
		if maxSimilarity < threshold {
			log.Println("not found")
			groups = append(groups, generateNewGroup(items[i], itemVector))
		} else {
			log.Println("found", maxGroup.GroupId)
			ic.Update(bson.M{"num_iid": items[i].NumIid}, bson.M{"$set": bson.M{"group_id": maxGroup.GroupId}})
			/*mergeVector(maxGroup.Vector, maxGroup.NumItem, itemVector)
			if CompareVector(itemVector, maxGroup.Vector) > CompareVector(maxGroup.DelegateVector, maxGroup.Vector) {
                fmt.Println("instead", itemVector,  maxGroup.DelegateVector)
				maxGroup.DelegateId = items[i].NumIid
				maxGroup.DelegateVector = itemVector
                fmt.Println("after", itemVector, maxGroup.DelegateVector)
			}
            */
            maxGroup.NumItem = maxGroup.NumItem + 1
            maxGroup.AveragePrice = (float32(maxGroup.NumItem)*maxGroup.AveragePrice + items[i].Price) / float32(maxGroup.NumItem+1)
			gc.Update(bson.M{"group_id": maxGroup.GroupId},
				bson.M{"$set": bson.M{
					"num_item":        maxGroup.NumItem,
					"average_price":   maxGroup.AveragePrice,
					//"vector":          maxGroup.Vector,
					//"delegate_id":     maxGroup.DelegateId,
					//"delegate_vector": maxGroup.DelegateVector,
				}})
		}
	}
    return lastStartTime
}

func main() {
	defer MgoSession.Close()
	for {
		startTime := getStartTime()
		log.Println("start", startTime)
		lastStartTime := scanTaobaoItems(startTime)
        saveStartTime(lastStartTime)
        log.Println("rest")
		time.Sleep(1 * time.Minute)
	}
}
