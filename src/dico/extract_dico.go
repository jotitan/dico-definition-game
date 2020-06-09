package dico

import (
	"encoding/json"
	"fmt"
	"github.com/jotitan/dico-definition-game/src/logger"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

func loadDicoFromDatasource(){
	words,waiter := runWordDefinitionSaver()
	waiter.Add(1)
	for i := 65 ; i < 91 ; i++ {
		findWordsAsync(string(rune(i)),1,waiter)
	}

	waiter.Done()
	logger.GetLogger2().Info("wait",len(words))
	waiter.Wait()
	data,_ := json.Marshal(words)
	ioutil.WriteFile("words.json",data,os.ModePerm)
}

func runWordDefinitionSaver()(map[string]string,*sync.WaitGroup){
	words := make(map[string]string,0)
	waiter := sync.WaitGroup{}
	go func(){
		for wordDefintion := range chanelDefinition {
			words[wordDefintion.Word] = wordDefintion.Definition
			waiter.Done()
		}
	}()
	return words,&waiter
}



func findWordsAsync(letter string,page int,waiter * sync.WaitGroup){
	url := filepath.Join(httpSource,"explore/def/" + letter)
	logger.GetLogger2().Info("Run page",letter,page)
	if page != 1 {
		url +=fmt.Sprintf("/%d", page)
	}
	if r,err := http.Get(url) ; err == nil {
		data, _ := ioutil.ReadAll(r.Body)
		reg, _ := regexp.Compile("(?:href=\"/definition/)([a-z-]+)(?:\">)([^<]*)")
		words := reg.FindAllStringSubmatch(string(data), -1)
		for _,result := range words {
			//For each word, search definition
			waiter.Add(1)
			searchWord(result[1],result[2])
		}
		if page == 1 {
			number := findNumberPage(letter,string(data))
			for i := 2 ; i <=number ; i++ {
				findWordsAsync(letter,i,waiter)
			}
		}
	}
}


func searchWord(urlWord,word string){
	go func(){
		definition := getWord(urlWord)
		chanelDefinition <- WordDefinition{word,definition}
		if strings.EqualFold("",definition) {
			logger.GetLogger2().Info(word,"has no definition")
		}
	}()
}

func findNumberPage(letter,data string)int{
	reg, _ := regexp.Compile("(?:href=\"/explore/def/" + letter + "/[0-9]+\">)([0-9]+)")
	results := reg.FindAllStringSubmatch(string(data), -1)
	if nb,err := strconv.ParseInt(results[len(results)-1][1],10,32);err == nil {
		return int(nb)
	}
	return 0
}

func getWord(word string)string{
	if r,err := http.Get(filepath.Join(httpSource,"definition/" + word)) ; err == nil {
		data, _ := ioutil.ReadAll(r.Body)
		dfn := extractField("d_dfn",data)
		if strings.EqualFold("",dfn){
			// extract d_xpl and concat to d_gls
			dfn = extractField("d_xpl",data) + extractField("d_gls",data)
		}
		return dfn
	}
	return ""
}

func extractField(field string,data []byte)string{
	reg, _ := regexp.Compile("(?:\"" + field + "\">)([^<]*)")
	result := reg.FindStringSubmatch(string(data))
	if len(result) == 0 {
		return ""
	}
	return result[1]
}
