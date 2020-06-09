package server

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jotitan/dico-definition-game/src/dico"
	"github.com/jotitan/dico-definition-game/src/logger"
	"net/http"
)

// Manage many games
type NetworkGameManager struct {
	// Key is the code to access
	games map[string]*NetworkGame
}

func NewNetworkManager()*NetworkGameManager {
	return &NetworkGameManager{make(map[string]*NetworkGame)}
}

func (m *NetworkGameManager)AddNewGame(dico dico.Dico)*NetworkGame{
	networkGame :=NewNetworkGame(dico)
	m.games[networkGame.game.Code] = networkGame
	return networkGame
}

func (m *NetworkGameManager)Get(gameCode string)*NetworkGame{
	if game,exist := m.games[gameCode] ; exist {
		return game
	}
	return nil
}

func (m NetworkGameManager)CheckCurrentPlayer(gameCode,player string)error{
	if ng,exist := m.games[gameCode] ; exist {
		if !ng.game.CheckCurrentDicoPlayer(player) {
			return errors.New("bad player")
		}
	}else{
		return errors.New("unknown game")
	}
	return nil
}

func (m *NetworkGameManager)Join(gameCode, playerName string, c *gin.Context)error{
	if ng,exist := m.games[gameCode] ; exist {
		return ng.Join(playerName,c)
	} else{
		return errors.New("unknown game")
	}
}

func writeEvent(w http.ResponseWriter, eventName,message string){
	logger.GetLogger2().Info("Write message",eventName,message)
	w.Write([]byte(fmt.Sprintf("event: %s\n",eventName)))
	w.Write([]byte("data: " + message + "\n\n"))
	w.(http.Flusher).Flush()
}