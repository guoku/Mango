package segment

import (
	"regexp"
	"strings"
)

type Node struct {
	Word     string
	Children []*Node
	Exist    bool //判断到这个节点是否构成一个词
	Origin   int  //指向这个词的根源的index,比如欧莱雅指向 olay/欧莱雅
}

type Text struct {
	Words string
	Freq  int
}
type TrieTree struct {
	MaxLength int
	Root      *Node
	Words     []*Text //原始的品牌名，比如olay/欧莱雅,我拆分为olay，欧莱雅，olay欧莱雅这三个构成前缀树，但最后提交的结果应该是olay/欧莱雅
	//Node里的origin就表示该node所表示的品牌对应的原始名字在Words里的index
}

func (this *TrieTree) Add(words string, freq int) {
	index := len(this.Words)
	text := &Text{Words: words, Freq: freq}
	this.Words = append(this.Words, text)
	if strings.Contains(words, "/") {
		arr := strings.Split(words, "/")
		first := this.clean(arr[0])
		second := this.clean(arr[1])
		third := this.clean(words)
		this.add(first, index)
		this.add(second, index)
		this.add(third, index)
	} else {
		this.add(words, index)
	}
}
func (this *TrieTree) add(words string, loc int) {
	current := this.Root
	if this.MaxLength < len(words) {
		this.MaxLength = len(words)
	}
	for _, word := range words {
		if current == nil {
			//树还没有初始化
			children := make([]*Node, 0, 50)
			current = &Node{Children: children}
			this.Root = current
		}
		nodes := current.Children
		var index int
		var exist bool
		if nodes == nil {
			nodes = make([]*Node, 0, 50)
		}
		index, exist = this.judge(nodes, string(word))
		if !exist {

			n := Node{Word: string(word), Exist: false}
			index = len(nodes)
			nodes = append(nodes, &n)
		}
		current.Children = nodes
		current = nodes[index]

	}
	current.Exist = true
	current.Origin = loc
}
func (this *TrieTree) clean(words string) string {
	re := regexp.MustCompile("\\pP|`| ")
	s := re.ReplaceAllLiteralString(words, "")
	return strings.ToLower(s)
}
func (this *TrieTree) Search(words string) (*Text, bool) {
	words = this.clean(words)
	current := this.Root
	for _, v := range words {
		if current == nil {
			return nil, false
		}
		nodes := current.Children
		index, exist := this.judge(nodes, string(v))
		if !exist {
			return nil, false
		}
		current = nodes[index]
	}
	return this.Words[current.Origin], current.Exist
}

func (this *TrieTree) Extract(text string, mode int) []string {
	//1表示最短模式，2表示最长模式，3表示全部结果都返回
}
func (this *TrieTree) judge(nodes []*Node, word string) (int, bool) {
	for k, v := range nodes {
		if word == v.Word {
			return k, true
		}
	}
	return -1, false
}
