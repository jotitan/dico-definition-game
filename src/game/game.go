package game

import (
	"errors"
	"math"
	"math/rand"
	"strings"
	"time"
)

const (
	StatusWaitingPlayers = Status("waiting_players")
	StatusWaitingRules   = Status("waiting_rules")
	StatusChoosingWord   = Status("choosing_word")
	StatusDefinition     = Status("definition")
	StatusVotes          = Status("votes")
)

const (
	WaitDefinition = 90000
	WaitVote = 60000
	WaitRules = 60000
)

type Status string

func (gs Status)Equals(gameStatus Status)bool{
	return strings.EqualFold(string(gs),string(gameStatus))
}

type Definition struct {
	definition string
	playerId string
	isCorrect bool
}

type Round struct {
	Word              string
	playersDefinition []Definition
	// The votes, for each player, position of definition
	votes map[string]int
}

func NewRound()*Round{
	return &Round{"",make([]Definition,0),make(map[string]int)}
}

func (r * Round)GetGoodDefinition()string{
	for _,def := range r.playersDefinition {
		if def.isCorrect {
			return def.definition
		}
	}
	return ""
}

func (r * Round) chooseWord(playerId,word, definition string){
	r.Word = word
	r.playersDefinition = append(r.playersDefinition,Definition{definition,playerId,true})
}

func (r * Round) addDefinition(playerId,definition string)error{
	// Check if player is not the master
	// Check if player has already send an answer
	if r.hasPlayerDefinition(playerId){
		return errors.New("player has already answered")
	}
	r.playersDefinition = append(r.playersDefinition,Definition{definition,playerId,false})
	return nil
}

func (r Round)hasPlayerDefinition(playerId string)bool{
	for _,def := range r.playersDefinition {
		if strings.EqualFold(def.playerId,playerId){
			return true
		}
	}
	return false
}

//countScore count score based on each vote. If vote for definition, two point for player, otherwise one point for definition creator. Each error, one point for master
// Return new score and point of round
func (r Round)countScore(players map[string]*Player, master *Player)(map[string]int,map[string][]int){
	roundScore := make(map[string]int,len(players))
	detailScore := make(map[string][]int,len(players))
	for _,p := range players {
		roundScore[p.Name] = 0
		detailScore[p.Name] = make([]int,0)
	}
	for playerID,vote := range r.votes {
		name := players[playerID].Name
		if r.playersDefinition[vote].isCorrect {
			players[playerID].Score+=2
			roundScore[name] += 2
			detailScore[name] = append(detailScore[name],2)
		}else{
			idDefinitionPlayer := r.playersDefinition[vote].playerId
			players[idDefinitionPlayer].Score++
			roundScore[players[idDefinitionPlayer].Name] ++
			detailScore[players[idDefinitionPlayer].Name] = append(detailScore[name],1)
			master.Score++
			roundScore[master.Name] ++
			detailScore[master.Name] = append(detailScore[master.Name],2)
		}
	}
	return roundScore,detailScore
}

type definitionWithInfo struct {
	Definition string
	IsPlayerAnswer bool
}

func (r * Round)GetDefinitionsWithInfo(playerID string)[]definitionWithInfo{
	definitions := make([]definitionWithInfo,len(r.playersDefinition))
	for i,def := range r.playersDefinition {
		definitions[i] = definitionWithInfo{def.definition,strings.EqualFold(def.playerId,playerID)}
	}
	return definitions
}

func (r *Round)GetDefinitions()[]string{
	definitions := make([]string,len(r.playersDefinition))
	for i,def := range r.playersDefinition {
		definitions[i] = def.definition
	}
	return definitions
}

func (r *Round) Vote(playerID string, definition int) error{
	if _,exist := r.votes[playerID]; exist {
		return errors.New("player as already voted")
	}
	r.votes[playerID] = definition
	return nil
}

func (r * Round)getVoters()[]string{
	voters := make([]string,0,len(r.votes))
	for id := range r.votes{
		voters = append(voters,id)
	}
	return voters
}

func (r * Round)getDefininers()[]string{
	defininers := make([]string,0,len(r.playersDefinition))
	for _,def := range r.playersDefinition{
		defininers = append(defininers,def.playerId)
	}
	return defininers
}

type Game struct {
	Code string
	// First player is the creator of the game
	Players []*Player
	playersById map[string]*Player
	// Player who search a Word in dico
	CurrentDicoPlayer int
	CurrentRound *Round
	// One of 5 status
	Status Status
	limitStepTime time.Time
	rules map[string]int
}

func Create()*Game{
	return &Game{Code:generateRandomCode(4),Status: StatusWaitingPlayers,
		Players:make([]*Player,0),
		playersById:make(map[string]*Player),
		rules:make(map[string]int),
		CurrentDicoPlayer:-1}
}

func (g * Game)DisconnectPlayer(playerID string)*Player{
	if player,exist := g.playersById[playerID];exist {
		player.Disconnect = true
		return player
	}
	return nil
}

func (g * Game)ReconnectPlayer(playerID string)error{
	if player,exist := g.playersById[playerID];exist {
		player.Disconnect = false
		return nil
	}
	return errors.New("unknown player")
}

func (g * Game) CheckStatus(status Status)bool{
	return g.Status.Equals(status)
}

func (g * Game)getRules()[]string{
	rules := make([]string,0,len(g.rules))
	for id := range g.rules{
		rules = append(rules,id)
	}
	return rules
}

func (g * Game)GetPlayerById(playerID string)(*Player,error) {
	if player,exist := g.playersById[playerID] ; exist {
		return player,nil
	}
	return nil,errors.New("unknown player")
}

func generateRandomCode(length int)string{
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := ""
	for i := 0 ; i < length ; i++ {
		code+=string(rune(r.Int()%26+65))
	}
	return code
}

func (g * Game)Join(name string)(*Player,error){
	if !g.CheckStatus(StatusWaitingPlayers) {
		return nil,errors.New("impossible to join game")
	}

	// Check existing name
	if g.isNameExist(name){
		return nil,errors.New("name already exists")
	}
	return g.addPlayer(name),nil
}

func (g * Game)isNameExist(name string)bool{
	for _,p := range g.Players {
		if strings.EqualFold(name,p.Name) {
			return true
		}
	}
	return false
}

func (g *Game)addPlayer(name string)*Player{
	player := NewPlayer(name)
	g.Players = append(g.Players,player)
	g.playersById[player.ID] = player
	return player
}

func (g * Game)Start(){
	g.NextRound()
}

func (g * Game)GetNbActivePlayers()int{
	nb := 0
	for _,player := range g.Players {
		if !player.Disconnect {
			nb++
		}
	}
	return nb
}


func (g *Game)NextRound(){
	// Change dico player
	g.CurrentDicoPlayer = (g.CurrentDicoPlayer+1)%len(g.Players)
	g.CurrentRound = NewRound()
	g.Status = StatusChoosingWord
}

func (g * Game)CheckCurrentDicoPlayer(player string)bool{
	return strings.EqualFold(player,g.Players[g.CurrentDicoPlayer].ID)
}

func (g * Game)GetCurrentDicoPlayer()*Player{
	return g.Players[g.CurrentDicoPlayer]
}

func (g *Game) ChooseWord(word,definition string) error{
	if !g.CheckStatus(StatusChoosingWord) {
		return errors.New("impossible to choose Word now")
	}

	g.CurrentRound.chooseWord(g.GetCurrentDicoPlayer().ID,word,definition)
	g.Status = StatusDefinition
	g.setLimitTime(WaitDefinition)
	return nil
}

func (g * Game)HasEverybodyAnswered()bool{
	return len(g.CurrentRound.playersDefinition) == len(g.Players)
}

func (g * Game)LaunchVotes(){
	g.Status = StatusVotes
	// Shuffle list of answer
	l := g.CurrentRound.playersDefinition
	rand.Shuffle(len(l),func(a,b int){l[a],l[b]=l[b],l[a]})

	g.setLimitTime(WaitVote)
}

func (g * Game)AddWordDefinition(playerId,definition string)error{
	if !g.CheckStatus(StatusDefinition) {
		return errors.New("impossible to give definition now")
	}
	if strings.EqualFold(playerId,g.GetCurrentDicoPlayer().ID) {
		return errors.New("master player can't give a definition")
	}
	return g.CurrentRound.addDefinition(playerId,definition)
}

func (g *Game) Vote(playerID string, definition int) error{
	if !g.CheckStatus(StatusVotes) {
		return errors.New("impossible to vote for definition now")
	}

	if strings.EqualFold(playerID,g.GetCurrentDicoPlayer().ID){
		return errors.New("master player can't vote")
	}
	return g.CurrentRound.Vote(playerID,definition)
}

// Return score of round and total score
func (g *Game)Count()(map[string]int,map[string][]int,map[string]int){
	roundScore,detailScore := g.CurrentRound.countScore(g.playersById,g.Players[g.CurrentDicoPlayer])
	return roundScore,detailScore,g.GetTotalScore()
}

func (g*Game)GetTotalScore()map[string]int{
	totalScore := make(map[string]int,len(g.playersById))
	for _,player := range g.playersById{
		totalScore[player.Name] = player.Score
	}
	return totalScore
}

func (g *Game) RulesReading() {
	g.Status = StatusWaitingRules
	g.setLimitTime(WaitRules)
}

func (g * Game)setLimitTime(wait int){
	g.limitStepTime = time.Now().Add(time.Duration(wait)*time.Millisecond)
}

func (g * Game)GetRestingTime()int{
	return int(math.Floor(g.limitStepTime.Sub(time.Now()).Seconds()))
}

func (g * Game)GetAnswers()[]string{
	// if definition step, return list of people who have answered
	if g.Status == StatusDefinition {
		return g.getPlayersNameFromID(g.CurrentRound.getDefininers())
	}
	if g.Status == StatusVotes {
		return g.getPlayersNameFromID(g.CurrentRound.getVoters())
	}
	if g.Status == StatusWaitingRules {
		return g.getPlayersNameFromID(g.getRules())
	}
	return []string{}
}

func (g * Game)getPlayersNameFromID(ids []string)[]string{
	players := make([]string,len(ids))
	for i,answer := range ids {
		players[i] = g.playersById[answer].Name
	}
	return players
}

func (g *Game) ReadRules(playerID string) *Player{
	if player,err := g.GetPlayerById(playerID) ; err == nil {
		g.rules[playerID] = 0
		return player
	}
	return nil
}