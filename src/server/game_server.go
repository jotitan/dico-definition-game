package server

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jotitan/dico-definition-game/src/dico"
	"github.com/jotitan/dico-definition-game/src/game"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type Server struct {
	dico dico.Dico
	gameManager *NetworkGameManager
	resources string
}

func NewGameServer(resources string)Server{
	return Server{dico.LoadDicoFromFile(), NewNetworkManager(),resources}
}

func parseInt(value string)int{
	if nb,err := strconv.ParseInt(value,10,32) ; err == nil {
		return int(nb)
	}
	return 0
}

func (s Server)GetWordsByLetter(c *gin.Context){
	letter := c.Param("letter")
	from := parseInt(c.Param("from"))
	to := parseInt(c.Param("to"))
	c.JSON(http.StatusOK,s.dico.GetWords(letter,from,to))
}

func (s Server)GetNbWordsByLetter(c *gin.Context){
	letter := c.Param("letter")
	c.JSON(http.StatusOK,gin.H{"nb":s.dico.Length(letter)})
}

// Extract game and player id
func (s Server)extractInfos(c *gin.Context)(*NetworkGame,string){
	if name,err := c.Cookie("player"); err == nil {
		if game := s.gameManager.Get(c.Param("game_code")) ; game != nil {
			return game,name
		}
	}
	return nil,""
}

func (s Server)StartGame(c *gin.Context){
	// Creator of game start the game when everybody is connected
	if game,player := s.extractInfos(c) ; game != nil {
		game.StartGame(c,player)
	}
}

func (s Server)GetGame(c *gin.Context){
	c.Header("Access-Control-Allow-Origin","*")
	if game := s.gameManager.Get(c.Param("game_code")) ; game != nil {
		c.JSON(http.StatusOK,gin.H{"nb":len(game.game.Players)})
	}else{
		c.AbortWithError(404,errors.New("game not found"))
	}
}

func (s Server)CreateGame(c *gin.Context){
	c.Header("Access-Control-Allow-Origin","*")
	ngame := s.gameManager.AddNewGame(s.dico)
	c.JSON(http.StatusOK,gin.H{"code":ngame.game.Code})
}

func (s Server)ReadRules(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	if game, player := s.extractInfos(c); game != nil {
		game.ReadRules(player)
	}
}

func (s Server)GiveDefinition(c *gin.Context){
	c.Header("Access-Control-Allow-Origin","*")
	if game,player := s.extractInfos(c) ; game != nil {
		values := s.extractParameters(c,[]string{"definition"})
		if err := game.GiveDefinition(player,values[0].(string)); err != nil {
			c.AbortWithError(404,err)
		}
	}
}

func (s Server)VoteDefinition(c *gin.Context){
	c.Header("Access-Control-Allow-Origin","*")
	if game,player := s.extractInfos(c) ; game != nil {
		values := s.extractParameters(c,[]string{"vote"})
		if err := game.VoteDefinition(player,int(values[0].(float64))); err != nil {
			c.AbortWithError(404,err)
		}
	}
}

// Choose word of user
func (s Server)ChooseWord(c *gin.Context){
	c.Header("Access-Control-Allow-Origin","*")
	values := s.extractParameters(c,[]string{"word"})
	if game,player := s.extractInfos(c) ; game != nil {
		if err := game.ChooseWord(values[0].(string),player) ; err != nil {
			c.AbortWithError(404,err)
		}
	}
}

func (s Server) Connect(c *gin.Context){
	c.Header("Access-Control-Allow-Origin","*")
	if game,player := s.extractInfos(c) ; game != nil {
		// If player already exist and reconnect (status disconnected), send context to him (score, current action)
		game.Connect(player,c)
	}
	c.AbortWithError(404,errors.New("impossible to find game or player"))
}

func (s Server)CanJoin(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	if ng := s.gameManager.Get(c.Param("game_code")) ; ng != nil {
		// Check if game is alreadt at waiting players status
		c.Writer.Write([]byte(fmt.Sprintf("{\"playable\":%t}",ng.game.Status == game.StatusWaitingPlayers)))
	}else{
		c.AbortWithError(404,errors.New("unknown game"))
	}
}
func (s Server)JoinGame(c *gin.Context){
	c.Header("Access-Control-Allow-Origin","*")
	if game,_:= s.extractInfos(c) ; game != nil {
		//game.Reconnect(player,c)
		//already joined, auto reconnect
	}
	values := s.extractParameters(c,[]string{"name"})
	if err := s.gameManager.Join(c.Param("game_code"),values[0].(string),c) ; err != nil {
		c.AbortWithError(404,err)
	}
}

func (s Server)extractParameters(c *gin.Context,fields []string)[]interface{} {
	m := make(map[string]interface{})
	c.BindJSON(&m)
	values := make([]interface{},len(fields))
	for i,field := range fields {
		values[i] = m[field]
	}
	return values
}

func (s Server)Default(c *gin.Context){
	url := c.Request.URL.Path[1:]
	if strings.HasPrefix(url,"static/") {
		http.ServeFile(c.Writer,c.Request,filepath.Join(s.resources,url))
		return
	}
	if filepath.Ext(url) != "" {
		http.ServeFile(c.Writer,c.Request,filepath.Join(s.resources,url))
		return
	}
	http.ServeFile(c.Writer,c.Request,filepath.Join(s.resources,"index.html"))
}

func (s Server)Run(){
	server := gin.Default()
	server.GET("/api/dico/search/:letter/:from/:to",s.GetWordsByLetter)
	server.GET("/api/dico/nb/:letter",s.GetNbWordsByLetter)
	server.POST("/api/game/create",s.CreateGame)
	server.GET("/api/game/detail/:game_code",s.GetGame)
	server.GET("/api/game/start/:game_code",s.StartGame)
	server.POST("/api/game/join/:game_code",s.JoinGame)
	server.GET("/api/game/can_join/:game_code",s.CanJoin)
	server.GET("/api/game/connect/:game_code",s.Connect)
	server.POST("/api/game/action/:game_code/:player")
	server.POST("/api/game/choose_word/:game_code",s.ChooseWord)
	server.POST("/api/game/vote/:game_code",s.VoteDefinition)
	server.POST("/api/game/give_definition/:game_code",s.GiveDefinition)
	server.POST("/api/game/read_rules/:game_code",s.ReadRules)
	server.NoRoute(s.Default)
	server.Run(":9011")
}

