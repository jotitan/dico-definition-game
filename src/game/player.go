package game

import (
	"fmt"
	"math/rand"
)

type Player struct {
	Name       string
	ID         string
	Score      int
	Disconnect bool
}

func NewPlayer(name string)*Player {
	return &Player{name,fmt.Sprintf("%d",rand.Int()%1000000),0,false}
}
