package filter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Entity struct {
	ID       int      `json:"entity_id"`
	Titles   []string `json:"item_titles"`
	Category string   `json:"category_title"`
}
type WordsText []byte
type Result struct {
	Id         int
	Title      string
	Brands     []string
	CleanTitle string
	Category   string
}

func LoadData(offset, count int) ([]*Result, error) {
	link := "http://114.113.154.47:8000/management/entity/without/title/sync/?offset=%d&count=%d"
	link = fmt.Sprintf(link, offset, count)
	resp, err := http.Get(link)
	var rs []*Result
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var entities []Entity
	err = json.Unmarshal(body, &entities)
	if err != nil {
		log.Println(err.Error())
		return rs, err
	}
	tree := new(TrieTree)
	tree.LoadDictionary("10.0.1.23", "words", "brands")
	tree.LoadBlackWords("10.0.1.23", "words", "dict_chi_eng")
	for _, ent := range entities {
		if len(ent.Titles) == 0 {
			continue
		}
		title := ent.Titles[0]
		//	log.Println("\n")
		//	log.Println("原始的标题：", title)
		brands := tree.Extract(title)
		//	log.Println("抽取出来的品牌名：", brands)
		title = tree.Cleanning(title)
		//	log.Println("清理的标题：", title)
		result := Result{Id: ent.ID, Title: ent.Titles[0], Brands: brands, CleanTitle: title, Category: ent.Category}
		rs = append(rs, &result)
	}
	return rs, nil
}
func ToHTML(data []*Result, name string) {
	html := `
<!DOCTYPE html PUBLIC "->
<html>
<head>
    <meta http-equiv="Content-Type" content="text/html;charset=utf-8">
</head>
<body>
    <div style="text-align:center;">
    <table border="3" style="margin:auto;width:80%;" >
        <tr>
            <td>ID</td><td>title</td><td>brands</td><td>select</td><td>category</td>
        </tr>`
	for k, v := range data {
		t1 := `
            <tr border="3" bgColor=%s>
                <td>%d</td>
                <td>%s<br><br>%s</td>
                
                <td>
        `
		if k%2 == 0 {
			t1 = fmt.Sprintf(t1, "#3c8dc4", v.Id, v.Title, v.CleanTitle)
		} else {
			t1 = fmt.Sprintf(t1, "#ccc", v.Id, v.Title, v.CleanTitle)
		}
		s := ""
		for _, b := range v.Brands {
			s = b + `<br>`
			t1 = t1 + s

		}
		selectbrand := Brandsprocess(v.Brands)
		re := regexp.MustCompile("[a-zA-Z]+")
		selectbrand = re.ReplaceAllStringFunc(selectbrand, toUpper)
		t1 = t1 + `</td><td>` + v.Category + `</td><td>` + selectbrand + `</td></tr>`

		html = html + t1
	}
	html = html + "</table></div></body></html>"
	ioutil.WriteFile(name, []byte(html), 0777)
}
func TextSliceToString(text []WordsText) []string {
	var output []string
	for _, word := range text {
		output = append(output, string(word))
	}
	return output
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
					output[currentWord] = text[alphanumericStart:current]
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
			output[currentWord] = text[alphanumericStart:current]
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

func toUpper(text string) string {
	t := []byte(text)
	if len(t) > 0 {
		if t[0] >= 'a' && t[0] <= 'z' {
			t[0] = t[0] - 'a' + 'A'
		}
	}
	return string(t)
}

func Brandsprocess(brands []string) string {
	if len(brands) == 0 {
		return ""
	}

	var result string
	maxlen := 0
	//先抽取英文名，如果有多个英文名，取最长的
	re := regexp.MustCompile("^[a-zA-Z\\pP ]+$")
	for _, brand := range brands {
		match := re.Match([]byte(brand))
		if match {
			if len(brand) > maxlen {
				result = brand
				maxlen = len(result)
			}

		} else {
			if strings.Contains(brand, "/") {
				t := strings.Split(brand, "/")
				for _, name := range t {
					if re.Match([]byte(name)) {
						if len(name) > maxlen {
							result = name
							log.Println(name)
							maxlen = len(name)
						}
					}
				}
			}
		}
	}
	if maxlen == 0 {
		return brands[0]
	}
	return result
}
