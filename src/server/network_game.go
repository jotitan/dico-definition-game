package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jotitan/dico-definition-game/src/dico"
	"github.com/jotitan/dico-definition-game/src/game"
	"github.com/jotitan/dico-definition-game/src/logger"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Special structure to manage timeout with players action. IF players answered too late, do next action
type requestsWaiter struct {
	waiter *sync.WaitGroup
}

func createRequestsWaiter(nb int, timeoutInMs int,nextAction func())*requestsWaiter{
	rw := requestsWaiter{&sync.WaitGroup{}}
	rw.waiter.Add(nb)
	chanWait := make(chan struct{},1)
	// When all responses receives, send message in chan
	go func(){
		rw.waiter.Wait()
		chanWait<-struct{}{}
	}()
	// Wait first chanel to respond
	go func() {
		select {
		case <-chanWait:
			fmt.Println("All responses receives")
			nextAction()
		case <-time.NewTimer(time.Duration(timeoutInMs) * time.Millisecond).C:
			fmt.Println("Timeout")
			nextAction()
		}
	}()
	return &rw
}

func (rw requestsWaiter)receive(){
	rw.waiter.Done()
}

type networkPlayer struct {
	sse chan message
	name string
	id string
}

func newNetworkPlayer(w http.ResponseWriter, r *http.Request, name,playerID string,watcher chan string)networkPlayer{
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create chanel to communicate with
	chanelEvent := make(chan message,10)

	// If connexion stop, close chanel
	np := networkPlayer{sse:chanelEvent, id:playerID,name:name}
	np.watchEndSSE(r,chanelEvent,watcher)
	return np
}

func (np networkPlayer) watchEndSSE(r * http.Request, chanelEvent chan message,watcher chan string){
	go func(){
		<- r.Context().Done()
		logger.GetLogger2().Info("Stop connexion")
		//remove player
		close(chanelEvent)
		watcher<-np.id
	}()
}

type message struct {
	event string
	data string
}

func watchMessage(w http.ResponseWriter, chanelEvent chan message){
	for {
		if message, more  := <- chanelEvent ; more {
			writeEvent(w,message.event,message.data)
		}else{
			break
		}
	}
}

// Manage all request received on rest API on a game. Wrap a game
type NetworkGame struct {
	game *game.Game
	// Key is id of player
	playersConnexion map[string]networkPlayer
	dico dico.Dico
	// Manage a waiter timeout for requests
	requestsWaiter *requestsWaiter
	// watch disconnect
	watcher chan string
}

func NewNetworkGame(typeGame string,dico dico.Dico)*NetworkGame {
	ng := &NetworkGame{game:game.Create(typeGame),dico:dico,playersConnexion:make(map[string]networkPlayer),watcher:make(chan string,10)}
	ng.watchDisconnect()
	return ng
}

func (ng * NetworkGame)watchDisconnect(){
	go func(){
		for{
			playerID := <- ng.watcher
			logger.GetLogger2().Info("Player", playerID,"Disconnect")
			delete(ng.playersConnexion, playerID)
			if player := ng.game.DisconnectPlayer(playerID) ; player != nil {
				ng.sendMessageToAll("-1",message{"notify",fmt.Sprintf("{\"type\":\"disconnect\",\"player\":\"%s\"}",player.Name)})
			}
		}
	}()
}

func (ng  * NetworkGame)addPlayer(playerID, name string,c *gin.Context)networkPlayer{
	np := newNetworkPlayer(c.Writer,c.Request,name,playerID,ng.watcher)
	ng.playersConnexion[playerID] = np
	return np
}

func (ng * NetworkGame)Connect(playerID string,c *gin.Context){
	if player,err := ng.game.GetPlayerById(playerID) ; err != nil {
		// Error
		logger.GetLogger2().Error("Got error",err)
	}else {
		np := ng.addPlayer(playerID, player.Name, c)
		if player.Disconnect {
			player.Disconnect = false
			np.sse <- ng.GetGameContext(playerID)
			np.sse <- ng.getSimpleScoreMessage()
			// Send to other that user is connected again
			ng.sendMessageToAll(playerID, message{"notify",fmt.Sprintf("{\"type\":\"reconnect\",\"player\":\"%s\"}", player.Name)})
		}else {
			ng.sendPlayerList(np)
			ng.sendMessageToAll(playerID, message{"message",fmt.Sprintf("{\"type\":\"welcome\",\"player\":\"%s\"}", player.Name)})
		}
		// Blocking
		watchMessage(c.Writer, np.sse)
	}
}

// Used to resend context to an disconnected player
// @param playerID used to pesonnalize message in some case
func (ng * NetworkGame)GetGameContext(playerID string)message{
	// Also send score
	switch ng.game.Status {
	case game.StatusWaitingPlayers:return ng.getMessagePlayers()
	case game.StatusChoosingWord:return ng.getMessageNewRound()
	case game.StatusWaitingRules:return ng.getMessageRules(true)
	case game.StatusDefinition:return ng.getMessageDefinition(ng.game.CurrentRound.Word,true)
	case game.StatusVotes:
		m,_ := ng.getMessageVotes(playerID,true)
		return m
	case game.StatusScore:
		return ng.getMessageScore(ng.game.ComputeCount())
	}
	return message{}
}

func (ng *NetworkGame)sendPlayerList(player networkPlayer){
	player.sse <- ng.getMessagePlayers()
}

func (ng * NetworkGame)createRequestWaiter(duration,nbAvoidAnswer int,fct func()){
	ng.requestsWaiter = createRequestsWaiter(ng.game.GetNbActivePlayers()-nbAvoidAnswer,duration,fct)
}

func (ng * NetworkGame)StartGame(c *gin.Context, playerID string){
	// Check if playerID is creator
	// Read rules ?
	ng.game.RulesReading()
	ng.sendMessageToAll("-1",ng.getSimpleScoreMessage())
	ng.sendMessageToAll("-1",ng.getMessageRules(false))
	ng.createRequestWaiter(game.WaitRules,0,func(){ng.StartRound()})
}

func (ng * NetworkGame)StartRound(){
	ng.game.Start()
	// Send current user
	m := ng.getMessageNewRound()
	ng.sendMessageToAll("-1",m)
}


func (ng * NetworkGame)Join(name string,c *gin.Context)error{
	if player,err := ng.game.Join(name) ; err != nil {
		return err
	}else{
		// Set cookie with
		c.SetCookie("player",player.ID,0,"/","",false,false)
		c.Writer.Write([]byte(fmt.Sprintf("{\"id\":\"%s\",\"type\":\"%s\"}",player.ID,ng.game.GetType())))
	}
	return nil
}

func (ng * NetworkGame)ReadRules(playerID string)error{
	if ng.requestsWaiter == nil || !ng.game.CheckStatus(game.StatusWaitingRules){
		// Impossible, must be in this status
		return errors.New("impossible to receive rules ack")
	}
	return ng.ackReadByPlayer(playerID)
}

func (ng * NetworkGame)ReadScore(playerID string)error{
	if ng.requestsWaiter == nil || !ng.game.CheckStatus(game.StatusScore){
		// Impossible, must be in this status
		return errors.New("impossible to receive rules ack")
	}
	return ng.ackReadByPlayer(playerID)
}

func (ng * NetworkGame)ackReadByPlayer(playerID string)error{
	player := ng.game.SetPlayerAnswer(playerID)
	ng.sendMessageToAll("-1",message{"notify",fmt.Sprintf("{\"type\":\"answer\",\"player\":\"%s\",\"countdown\":%d}",player.Name,ng.game.GetRestingTime(),)})
	ng.requestsWaiter.receive()
	return nil
}


func (ng * NetworkGame)ChooseWord(word, playerID string)error{
	if !ng.game.CheckCurrentDicoPlayer(playerID) {
		return errors.New("this player can't choose word")
	}
	if definition,err := ng.dico.GetDefinition(word); err != nil {
		return err
	}else{
		if err := ng.game.ChooseWord(word,definition);err != nil {
			return err
		}
		// Send message to user for giving : message + receiver
		ng.sendMessageToAll("-1",ng.getMessageDefinition(word,false))
		ng.createRequestWaiter(game.WaitDefinition,ng.getNumberAvoid(),func(){ng.StartVotes()})
	}
	return nil
}

func (ng * NetworkGame)getNumberAvoid()int{
	if !ng.game.TypeGameNormal{
		return 0
	}
	return 1
}

func formatPlayer(p * game.Player)string{
	return fmt.Sprintf("{\"id\":\"%s\",\"name\":\"%s\"}",p.ID,p.Name)
}

// if answers, add list a people who answered
func (ng * NetworkGame) getMessageDefinition(word string,answers bool)message{
	answersData := ng.getAnswersData(answers)
	return message{"message",fmt.Sprintf("{\"type\":\"definition\",\"word\":\"%s\",\"countdown\":%d,\"master\":%s%s}",word,
		ng.game.GetRestingTime(),
		formatPlayer(ng.game.GetCurrentDicoPlayer()),
		answersData)}
}

func (ng * NetworkGame)getMessageScore(roundScore map[string]int,detailScore map[string]*game.DetailScore,totalScore map[string]int)message{
	answersData := ng.getAnswersData(true)
	goodAnswer := ng.game.CurrentRound.GetGoodDefinition()
	word := ng.game.CurrentRound.Word
	roundData,_ := json.Marshal(roundScore)
	totalData,_ := json.Marshal(totalScore)
	detailData,_ := json.Marshal(detailScore)
	definitions,_ := json.Marshal(ng.game.GetDefinitionsWithName())
	data := fmt.Sprintf("{\"type\":\"score\",\"word\":\"%s\",\"answer\":\"%s\",\"round\":%s,\"total\":%s,\"detail\":%s,\"countdown\":%d%s,\"definitions\":%s}",
		word,goodAnswer,roundData,totalData,detailData,ng.game.GetRestingTime(),answersData,definitions)
	return message{"message",data}
}

// playerID is used to specify which is the answer of plater
func (ng * NetworkGame)getMessageVotes(playerID string,answers bool)(message,error){

	// If fun, hide the good answer
	if data,err := json.Marshal(ng.game.CurrentRound.GetDefinitionsWithInfo(playerID,"-1")) ; err == nil {
		answersData := ng.getAnswersData(answers)
		return message{"message", fmt.Sprintf("{\"type\":\"vote\",\"master\":%s,\"countdown\":%d,\"word\":\"%s\",\"definitions\":%s%s}",
			formatPlayer(ng.game.GetCurrentDicoPlayer()),
			ng.game.GetRestingTime(),
			ng.game.CurrentRound.Word,
			string(data),
			answersData)},nil
	}
	return message{},errors.New("impossible to create event")
}

func (ng * NetworkGame)getAnswersData(answers bool)string{
	if answers {
		if data,err := json.Marshal(ng.game.GetAnswers()) ; err == nil {
			return fmt.Sprintf(",\"answers\":%s",string(data))
		}
	}
	return ""
}

func (ng * NetworkGame)StartVotes(){
	ng.game.LaunchVotes()
	// Send message for vote
	ng.sendMessageFromFctToAll("-1",func(playerID string)(message,error){return ng.getMessageVotes(playerID,false)})
	ng.createRequestWaiter(game.WaitVote, ng.getNumberAvoid(),func() { ng.ComputeScore() })
}

func (ng * NetworkGame)GiveDefinition(playerID,definition string)error{
	if ng.requestsWaiter == nil {
		return errors.New("impossible to give definition (waiter missing)")
	}
	if err :=  ng.game.AddWordDefinition(playerID,definition); err != nil {
		return err
	}
	player,_ :=ng.game.GetPlayerById(playerID)
	ng.sendMessageToAll("-1",message{"notify",fmt.Sprintf("{\"type\":\"answer\",\"player\":\"%s\",\"countdown\":%d}",player.Name,ng.game.GetRestingTime(),)})
	ng.requestsWaiter.receive()
	return nil
}

func (ng * NetworkGame)VoteDefinition(playerID string,definition int)error{
	if ng.requestsWaiter == nil {
		return errors.New("impossible to vote for definition (waiter missing)")
	}
	if err := ng.game.Vote(playerID,definition) ; err != nil {
		return err
	}
	player,_ :=ng.game.GetPlayerById(playerID)
	ng.sendMessageToAll("-1",message{"notify",fmt.Sprintf("{\"type\":\"answer\",\"player\":\"%s\",\"countdown\":%d}",player.Name,ng.game.GetRestingTime(),)})
	ng.requestsWaiter.receive()
	return nil
}

func (ng *NetworkGame) sendMessageToAll(playerID string,m message) {
	for id,player := range ng.playersConnexion{
		if !strings.EqualFold(id,playerID) {
			player.sse<-m
		}
	}
}

// Use to personnalize message for each user
func (ng *NetworkGame) sendMessageFromFctToAll(playerID string,fct func(string)(message,error)) {
	for id,player := range ng.playersConnexion{
		if !strings.EqualFold(id,playerID) {
			if m,err := fct(id) ; err == nil {
				player.sse<- m
			}
		}
	}
}

func (ng *NetworkGame) ComputeScore() {
	ng.game.SetScoreStatus()
	m := ng.getMessageScore(ng.game.Count())
	ng.sendMessageToAll("-1",m)
	ng.createRequestWaiter(game.WaitScoreReading, 0,func() { ng.StartRound() })
}

func (ng * NetworkGame)getMessageRules(answers bool)message{
	answersData := ng.getAnswersData(answers)
	return message{"message",fmt.Sprintf("{\"type\":\"rules\",\"countdown\":%d%s}",ng.game.GetRestingTime(),answersData)}
}

func (ng * NetworkGame)getMessageNewRound()message{
	return message{"message",fmt.Sprintf("{\"type\":\"round\",\"master\":%s}",formatPlayer(ng.game.GetCurrentDicoPlayer()))}
}

func (ng * NetworkGame)getMessagePlayers()message{
	players :=ng.getPlayerListAsString()
	return message{"message",fmt.Sprintf("{\"type\":\"players\",\"players\":%s}",players)}
}
func (ng *NetworkGame)getPlayerListAsString()string {
	names := make([]string, 0, len(ng.playersConnexion))
	for _, player := range ng.playersConnexion {
		names = append(names, player.name)
	}
	data,_ := json.Marshal(names)
	return string(data)
}

func (ng * NetworkGame)getSimpleScoreMessage()message{
	totalScore := ng.game.GetTotalScore()
	totalData,_ := json.Marshal(totalScore)
	return message{"notify",fmt.Sprintf("{\"type\":\"current-score\",\"total\":%s}",totalData)}
}