package game

import "testing"

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
