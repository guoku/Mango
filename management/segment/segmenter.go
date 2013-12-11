package segment

import (
	"github.com/qiniu/log"
	"regexp"
	"sort"
	"strings"
)

type Node struct {
	Word        string
	Black       string
	Normal      string
	Children    []*Node
	Exist       bool //判断到这个节点是否构成一个词
	BlackExist  bool
	NormalExist bool
	BlackOrigin int
	Origin      []int //指向这个词的根源的index,比如欧莱雅指向 olay/欧莱雅
}

type Text struct {
	Words string
	Freq  int
}
type TrieTree struct {
	MaxLength  int
	Root       *Node
	BlackWords []string
	Words      []Text //原始的品牌名，比如olay/欧莱雅,我拆分为olay，欧莱雅，olay欧莱雅这三个构成前缀树，但最后提交的结果应该是olay/欧莱雅
	//Node里的origin就表示该node所表示的品牌对应的原始名字在Words里的index
}

func (this *TrieTree) AddNormal(words string) {
	//添加普通词库
	words = strings.ToLower(words)
	slicewords := SplitTextToWords([]byte(words))
	texts := TextSliceToString(slicewords)
	current := this.Root
	if current == nil {
		children := make([]*Node, 0)
		current = &Node{Children: children}
		this.Root = current
	}
	for _, word := range texts {
		nodes := current.Children
		if nodes == nil {
			nodes = make([]*Node, 0)
		}
		index, exist := this.judgeAll(nodes, word)
		if !exist {
			n := Node{Normal: word, Exist: false}
			index = len(nodes)
			nodes = append(nodes, &n)
		} else {
			nodes[index].Normal = word

		}
		current.Children = nodes
		current = nodes[index]
	}
	current.NormalExist = true

}

func (this *TrieTree) judgeAll(nodes []*Node, word string) (int, bool) {
	for k, v := range nodes {
		if word == v.Normal || word == v.Black || word == v.Word {
			return k, true
		}
	}
	return -1, false
}
func (this *TrieTree) judgeNormal(nodes []*Node, word string) (int, bool) {
	for k, v := range nodes {
		if word == v.Normal {
			return k, true
		}
	}
	return -1, false
}
func (this *TrieTree) Add(words string, freq int) {
	//添加品牌词
	index := len(this.Words)
	text := Text{Words: words, Freq: freq}
	this.Words = append(this.Words, text)
	if strings.Contains(words, "/") {
		arr := strings.Split(words, "/")
		first := this.clean(arr[0])
		second := this.clean(arr[1])
		third := this.clean(words)
		if len(first) > 2 {
			this.add(first, index)
		}
		if len(second) > 2 {
			this.add(second, index)
		}
		if len(third) > 2 {
			this.add(third, index)
		}
	} else {
		if len(words) > 2 {
			this.add(words, index)
		}
	}
}
func (this *TrieTree) add(words string, loc int) {
	words = strings.ToLower(words)
	current := this.Root
	slicewords := SplitTextToWords([]byte(words))
	texts := TextSliceToString(slicewords)
	if this.MaxLength < len(texts) {
		this.MaxLength = len(texts)
	}
	if current == nil {
		//树还没有初始化
		children := make([]*Node, 0, 50)
		current = &Node{Children: children}
		this.Root = current
	}
	for _, word := range texts {
		nodes := current.Children
		var index int
		var exist bool
		if nodes == nil {
			nodes = make([]*Node, 0, 50)
		}
		index, exist = this.judgeAll(nodes, word)
		if !exist {
			n := Node{Word: word, Exist: false}
			index = len(nodes)
			nodes = append(nodes, &n)
		} else {
			nodes[index].Word = word
		}
		current.Children = nodes
		current = nodes[index]

	}
	current.Exist = true
	current.Origin = append(current.Origin, loc)
}
func (this *TrieTree) AddBlackWord(words string) {
	//添加垃圾词
	words = strings.ToLower(words)
	slicewords := SplitTextToWords([]byte(words))
	texts := TextSliceToString(slicewords)
	current := this.Root
	if current == nil {
		children := make([]*Node, 0)
		current = &Node{Children: children}
		this.Root = current
	}
	this.BlackWords = append(this.BlackWords, words)
	for _, word := range texts {
		nodes := current.Children
		if nodes == nil {
			nodes = make([]*Node, 0)
		}
		index, exist := this.judgeAll(nodes, word)
		if !exist {
			n := Node{Black: word, Exist: false}
			index = len(nodes)
			nodes = append(nodes, &n)
		} else {
			nodes[index].Black = word
		}
		current.Children = nodes
		current = nodes[index]
	}
	current.BlackExist = true
	current.BlackOrigin = len(this.BlackWords) - 1
}

func (this *TrieTree) Cleanning(title string) string {
	//根据黑名单，对标题进行清理
	//在查找黑名单的路径上，如果有品牌名，则停止不对其进行清理
	//如果找到一个要去掉的词，应该看其后继是否还存在品牌名
	//title = strings.ToLower(title)
	re := regexp.MustCompile("(&[a-z0-9]*;([a-z0-9];)?)|(【)|(】)|★|!|(<>)|(。)|(___)|(\\(\\))|(◆)|(\\*)|(\\p{S})|(（)|(）)|(满\\d+包邮)")
	title = re.ReplaceAllString(title, " ")
	slicewords := SplitTextToWords([]byte(title))
	texts := TextSliceToString(slicewords)
	current := this.Root
	passed := false
	var hit int
	var start int = 0
	for i := 0; i < len(texts); i++ {
		nodes := current.Children
		bi, be := this.judge(nodes, strings.ToLower(texts[i]), false)
		if be {
			if nodes[bi].Exist {
				current = this.Root
				continue
			}
		}
		//过期，效果不好，故不添加
		//对于普通词语，如果找到了一个垃圾词，还应该看这个垃圾词与其后的句子是否还构成有词
		//比如天然是垃圾词，但是天然石则不是垃圾词，故含有天然石的句子，不能够删除掉天然二字
		index, exist := this.judge(nodes, strings.ToLower(texts[i]), true)
		if !exist {
			if passed && current.Word == "" {
				if i == 0 {
					texts = texts[1:]
				} else {
					if start > 0 {
						start = start + 1
					}
					texts = append(texts[:start], texts[hit+1:]...)
				}
				i = 0

			}
			start = i
			current = this.Root
			passed = false
			continue
		}
		current = nodes[index]
		if current.BlackExist {
			hit = i
			log.Info(this.BlackWords[current.BlackOrigin])
			passed = true
		}

	}
	re = regexp.MustCompile("^[a-zA-Z].*[a-zA-Z0-9]$")
	for i := 0; i < len(texts); i++ {
		if re.MatchString(texts[i]) {
			texts[i] = texts[i] + " "
		}
	}
	title = strings.Join(texts, "")
	title = strings.TrimSpace(title)
	re = regexp.MustCompile(" +")
	title = re.ReplaceAllString(title, " ")
	/*
		ta := strings.Split(title, " ")
		for i := 0; i < len(ta); i++ {
			r := []rune(strings.TrimSpace(ta[i]))
			if len(r) == 1 {
				//单个字，需要删除
				if i < len(ta)-1 {
					ta = append(ta[:i], ta[i+1:]...)
				} else {
					ta = ta[:i]
				}
				i = i - 1
			}
		}
		title = strings.Join(ta, " ")
	*/
	return title
}

func (this *TrieTree) findNorm(words string) bool {
	//找到的垃圾词，与其后面紧接着的一个字，做匹配判断

	slicewords := SplitTextToWords([]byte(words))
	texts := TextSliceToString(slicewords)
	for i := len(texts) - 2; i >= 0; i-- {
		current := this.Root
		for _, v := range texts[i:] {
			if current == nil {
				return false
			}
			nodes := current.Children
			_, exist := this.judgeNormal(nodes, v)
			if exist {
				return true
			} else {
				break
			}
		}
	}
	return false
}
func (this *TrieTree) clean(words string) string {
	re := regexp.MustCompile("\\pP|`")
	s := re.ReplaceAllLiteralString(words, "")
	return strings.ToLower(s)
}
func (this *TrieTree) Search(words string) ([]Text, bool) {
	words = this.clean(words)
	slicewords := SplitTextToWords([]byte(words))
	texts := TextSliceToString(slicewords)
	current := this.Root
	var result []Text
	for _, v := range texts {
		if current == nil {
			return result, false
		}
		nodes := current.Children
		index, exist := this.judge(nodes, v, false)
		if !exist {
			return result, false
		}
		current = nodes[index]
	}
	var ts []Text
	for _, v := range current.Origin {
		ts = append(ts, this.Words[v])
	}
	return ts, current.Exist
}

func (this *TrieTree) Extract(text string) []string {
	text = this.clean(text)
	slicewords := SplitTextToWords([]byte(text))
	texts := TextSliceToString(slicewords)
	current := this.Root
	var result []Text
	var keys map[int]bool = make(map[int]bool)
	var flag bool = false
	var hit *Node
	for i := 0; i < len(texts); i++ {
		nodes := current.Children
		index, exist := this.judge(nodes, texts[i], false)
		if !exist {
			if flag {
				for _, v := range hit.Origin {
					keys[v] = true
				}
				flag = false
				i = i - 1
			}
			current = this.Root
			continue
		}
		current = nodes[index]
		if current.Exist {
			flag = true
			hit = current
		}
	}
	for key, _ := range keys {
		result = append(result, this.Words[key])
	}
	sort.Sort(ByFreq(result))
	var tmp []string
	for i := len(result) - 1; i >= 0; i-- {
		tmp = append(tmp, result[i].Words)
	}
	return tmp
}
func (this *TrieTree) judge(nodes []*Node, word string, black bool) (int, bool) {
	if black == false {
		for k, v := range nodes {
			if word == v.Word {
				return k, true
			}

		}
	} else {
		for k, v := range nodes {
			if word == v.Black {
				return k, true
			}
		}
	}
	return -1, false
}

//排序接口
type ByWord []*Node

func (this ByWord) Len() int {
	return len(this)
}
func (this ByWord) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}
func (this ByWord) Less(i, j int) bool {
	return this[i].Word < this[j].Word
}

type ByFreq []Text

func (this ByFreq) Len() int {
	return len(this)
}
func (this ByFreq) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}
func (this ByFreq) Less(i, j int) bool {
	return this[i].Freq < this[j].Freq
}
