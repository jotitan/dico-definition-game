package game

import (
	"fmt"
	"testing"
)

func TestGame(t *testing.T){
	game := Create()
	if game == nil {
		t.Error("Game must not be null")
	}
	if _,err := game.Join("robert") ; err != nil {
		t.Error("First player should be added")
	}

	if _,err := game.Join("robert") ; err == nil {
		t.Error("Robert must be used one")
	}
	game.Join("james")
	game.Join("bernard")
	if len(game.Players) != 3 {
		t.Error("Must find 3 players")
	}

	game.Start()
	if game.CurrentDicoPlayer != 0 {
		t.Error("First player must be 0")
	}

	game.ChooseWord("éclipse","Cache le soleil")
	if game.CurrentRound.addDefinition(game.Players[0].ID,"pas le droit") == nil {
		t.Error("Master can't give a definition")
	}
	if game.CurrentRound.addDefinition(game.Players[1].ID,"petit chaperon rouge") != nil {
		t.Error("Add definition can't generation an error")
	}
	if game.CurrentRound.addDefinition(game.Players[1].ID,"autre definition") == nil {
		t.Error("A player can't answer twice")
	}
	if game.HasEverybodyAnswered() {
		t.Error("All player must answered")
	}
	game.CurrentRound.addDefinition(game.Players[2].ID,"derniere definition")
	if !game.HasEverybodyAnswered() {
		t.Error("All player have answered")
	}
	game.LaunchVotes()
	goodAnswer := 0
	badAnswer := 0
	for i,def := range game.CurrentRound.playersDefinition{
		if def.isCorrect {
			goodAnswer = i
		}else{
			badAnswer = i
		}
	}
	if game.Vote(game.Players[0].ID,0) == nil {
		t.Error("Master player can't vote")
	}

	if err := game.Vote(game.Players[1].ID,goodAnswer) ;err!= nil {
		t.Error("Player should vote",err)
	}

	if game.Vote(game.Players[1].ID,1) == nil {
		t.Error("Player can't vote twice")
	}
	if err := game.Vote(game.Players[2].ID,badAnswer) ; err!= nil {
		t.Error("Player should vote",err)
	}
	game.Count()
	if game.Players[0].Score != 1 {
		t.Error("Player 0 should have 1 point")
	}
	if game.Players[1].Score != 3 {
		t.Error("Player 1 should have 3 points")
	}
	if game.Players[2].Score != 0 {
		t.Error("Player should have 0 point")
	}
}

func TestRound(t *testing.T){
	players := make(map[string]*Player)
	players["1"] = NewPlayer("P1")
	players["2"] = NewPlayer("P2")
	players["3"] = NewPlayer("P3")
	players["4"] = NewPlayer("P4")

	r := NewRound()
	r.chooseWord("1","veni vidi vici","Super")

	if r.addDefinition("2","def 2") != nil {
		t.Error("Player 2 can vote")
	}
	if r.addDefinition("3","def 3") != nil {
		t.Error("Player 3 can vote")
	}
	if r.addDefinition("4","def 4") != nil {
		t.Error("Player 4 can vote")
	}
	// No shuffle, good answer is first
	r.Vote("2",0)
	// 3 and 4 vote for def player 2
	r.Vote("3",1)
	r.Vote("4",1)
	// Results should be 2 , 4 , 0 , 0
	roundScore, detailScore := r.countScore(players,players["1"])
	if pt := roundScore["P1"] ; pt != 2 {
		t.Error(fmt.Sprintf("Player 1 should have 2 points but get %d",pt))
	}
	if pt := roundScore["P2"] ; pt != 4 {
		t.Error(fmt.Sprintf("Player 2 should have 4 points but get %d",pt))
	}
	if pt := roundScore["P3"] ; pt != 0 {
		t.Error(fmt.Sprintf("Player 3 should have 0 points but get %d",pt))
	}
	if pt := roundScore["P4"] ; pt != 0 {
		t.Error(fmt.Sprintf("Player 4 should have 0 points but get %d",pt))
	}
	if pt := detailScore["P1"].ErrorPoint ; pt!= 2 {
		t.Error(fmt.Sprintf("Master should win 2 points on error but found %d",pt))
	}
	if pt := detailScore["P2"].GoodDef; pt!=true {
		t.Error("P2 should have found the good answer")
	}
	if pt := detailScore["P2"].VotePoint; pt!=2{
		t.Error(fmt.Sprintf("P2 should win 2 points on by votes %d",pt))
	}
	if pt := detailScore["P3"].VotePoint; pt!=0{
		t.Error(fmt.Sprintf("P3 should win 0 points on by votes %d",pt))
	}
	if pt := detailScore["P4"].VotePoint; pt!=0{
		t.Error(fmt.Sprintf("P4 should win 0 points on by votes %d",pt))
	}
}
