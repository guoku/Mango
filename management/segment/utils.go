package segment

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"unicode"
	"unicode/utf8"
)

type Entity struct {
	ID     int      `json:"entity_id"`
	Titles []string `json:"item_titles"`
}
type WordsText []byte

func LoadData() {
	resp, err := http.Get("http://10.0.1.109:8000/management/entity/without/title/sync/")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var entities []Entity
	err = json.Unmarshal(body, &entities)
	if err != nil {
		log.Println(err.Error())
		return
	}
	tree := new(TrieTree)
	tree.LoadDictionary("10.0.1.23", "words", "brands")
	tree.LoadBlackWords("10.0.1.23", "words", "dict_chi_eng")
	for _, ent := range entities {
		if len(ent.Titles) == 0 {
			continue
		}
		title := ent.Titles[0]
		log.Println("原始的标题：", title)
		title = tree.Cleanning(title)
		log.Println("清理后的标题：", title)
		brands := tree.Extract(title)
		log.Println("抽取出来的品牌名：", brands)
	}
}
func SplitTextToWords(text WordsText) []WordsText {
	output := make([]WordsText, len(text))
	current := 0
	currentWord := 0
	inAlphanumeric := true
	alphanumericStart := 0
	for current < len(text) {
		r, size := utf8.DecodeRune(text[current:])
		if size <= 2 && (unicode.IsLetter(r) || unicode.IsNumber(r)) {
			// 当前是拉丁字母或数字（非中日韩文字）
			if !inAlphanumeric {
				alphanumericStart = current
				inAlphanumeric = true
			}
		} else {
			if inAlphanumeric {
				inAlphanumeric = false
				if current != 0 {
					output[currentWord] = toLower(text[alphanumericStart:current])
					currentWord++
				}
			}
			if text[current : current+size][0] != 32 {
				output[currentWord] = text[current : current+size]

				currentWord++

			}
		}
		current += size
	}

	// 处理最后一个字元是英文的情况
	if inAlphanumeric {
		if current != 0 {
			output[currentWord] = toLower(text[alphanumericStart:current])
			currentWord++
		}
	}

	return output[:currentWord]
}

// 将英文词转化为小写
func toLower(text []byte) []byte {
	output := make([]byte, len(text))
	for i, t := range text {
		if t >= 'A' && t <= 'Z' {
			output[i] = t - 'A' + 'a'
		} else {
			output[i] = t
		}
	}
	return output
}

func TextSliceToString(text []WordsText) []string {
	var output []string
	for _, word := range text {
		output = append(output, string(word))
	}
	return output
}
