package dico

import (
	"encoding/json"
	"errors"
	"github.com/jotitan/dico-definition-game/src/logger"
	"io/ioutil"
	"math"
	"math/rand"
	"sort"
	"strings"
)

const (
	httpSource = "https://dictionnaire.lerobert.com"
)

type Dico struct {
	byLetter map[string]WordsDefinition
	byWord map[string]*WordDefinition
}

func LoadDicoFromFile()Dico{
	if data,err := ioutil.ReadFile("words.json") ; err == nil {
		dico := make(map[string]string)
		json.Unmarshal(data,&dico)
		logger.GetLogger2().Info("Dico well loaded with",len(dico),"words")
		return newDico(dico)
	}
	return Dico{}
}

// Return a letter and a page
func (d Dico)GetRandomPage(nbByPage int)(string,int){
	value := string(65 + rand.Int()%26)
	nbWord := d.Length(value)
	nbPage := int(math.Max(1,math.Ceil(float64(nbWord/nbByPage))))
	randomPage := rand.Int()%nbPage
	return value,randomPage
}

func (d Dico)Length(letter string)int{
	return len(d.byLetter[letter[0:1]])
}

func (d Dico)GetDefinition(word string)(string,error){
	if wordDef,exist := d.byWord[word] ; exist {
		return wordDef.Definition,nil
	}
	return "",errors.New("word has no definition")
}

func (d Dico)GetWords(letter string,from,to int)[]*WordDefinition {
	words := d.byLetter[letter[0:1]]
	if from >=len(words){
		return []*WordDefinition{}
	}
	to = int(math.Min(float64(to),float64(len(words)-1)))
	return words[from:to]
}

func newDico(rawDico map[string]string)Dico{
	dicoByLetter := make(map[string]WordsDefinition)
	dicoByWord := make(map[string]*WordDefinition)
	for i := 65 ; i < 91 ; i++ {
		dicoByLetter[string(rune(i))] = make(WordsDefinition,0)
	}
	for word,definition := range rawDico {
		letter := strings.ToUpper(word[0:1])
		wordDef := &WordDefinition{word,definition}
		dicoByLetter[letter] = append(dicoByLetter[letter],wordDef)
		dicoByWord[word] = wordDef
	}
	for i := 65 ; i < 91 ; i++ {
		sort.Sort(dicoByLetter[string(rune(i))])
	}
	return Dico{byLetter:dicoByLetter,byWord:dicoByWord}
}

var chanelDefinition = make(chan WordDefinition,100)

type WordDefinition struct{
	Word string
	Definition string
}

type WordsDefinition []*WordDefinition

func (wd WordsDefinition)Len() int{return len(wd)}
func (wd WordsDefinition)Less(i, j int) bool{return wd[i].Word < wd[j].Word}
func (wd WordsDefinition)Swap(i, j int){wd[i],wd[j]=wd[j],wd[i]}

