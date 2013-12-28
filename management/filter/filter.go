package filter

import (
	"github.com/qiniu/log"
	"regexp"
	"strings"
)

//从分词的数组中清理垃圾词
func (this *TrieTree) Filtrate(texts [][]string) string {
	result := this.FiltrateForArray(texts)
	returnresult := sliceToString(result)
	re := regexp.MustCompile("[a-zA-z]+")
	returnresult = re.ReplaceAllStringFunc(returnresult, ToUpper)
	return returnresult
}

func (this *TrieTree) FiltrateForArray(texts [][]string) []string {
	for i := 0; i < len(texts); i++ {
		current := this.Root
		term := texts[i]
		hit := false
		for _, word := range term {
			nodes := current.Children
			index, exist := this.judge(nodes, word, true)
			if !exist {
				hit = false
				break
			}
			current = nodes[index]
			if current.BlackExist {
				log.Info(current.Black)
				hit = true
			} else {
				hit = false
			}
		}
		if hit {
			texts = append(texts[:i], texts[i+1:]...)
			//log.Info(texts)
			hit = false
			i = i - 1
		}
	}
	//log.Info(texts)
	var result []string
	for _, word := range texts {
		result = append(result, sliceToString(word))
	}
	return result
}

//从分词后的数组中提取品牌
func (this *TrieTree) FilterBrand(texts [][]string) string {
	for _, term := range texts {
		current := this.Root
		hit := false
		for _, word := range term {
			nodes := current.Children
			index, exist := this.judge(nodes, word, false)
			if exist {
				current = nodes[index]
				if current.Exist {
					hit = true
				} else {
					hit = false
				}
				continue
			} else {
				hit = false
				break
			}

		}
		if hit {
			result := sliceToString(term)
			re := regexp.MustCompile("[a-zA-z]+")
			result = re.ReplaceAllStringFunc(result, ToUpper)
			result = strings.Replace(result, "——", "-", -1)
			return result
		}
	}
	return ""
}

func sliceToString(text []string) string {
	re := regexp.MustCompile("^[a-zA-Z0-9 \\pP\\pS]+$")
	result := ""
	isAlpha := true
	rse := regexp.MustCompile("^[a-zA-Z0-9]")
	rse2 := regexp.MustCompile("[a-zA-Z0-9]$")
	for _, word := range text {
		word = strings.TrimSpace(word)
		match := re.MatchString(word)
		if match {
			if isAlpha {
				result = result + word + " "
			} else {
				result = result + " " + word + " "
				isAlpha = true
			}
		} else {
			if rse.MatchString(word) {
				result = result + " " + word
			} else if rse2.MatchString(word) {
				result = result + word + " "
			} else {
				result = result + word
			}
			isAlpha = false
		}
	}
	re = regexp.MustCompile(" (\\pP|\\pS) ")
	var repfunc = func(repl string) string {
		return strings.TrimSpace(repl)
	}
	result = re.ReplaceAllStringFunc(result, repfunc)
	return result
}
