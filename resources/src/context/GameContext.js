import React from "react";

export default React.createContext({
    createGame:type=>{},
    createCurrentGame:(code,isCreator)=>{},
    sendWordForGame:(code,word)=>{},
    isGameExist:code=>{},
    sendDefinitionForWord:(code,definition)=>{},
    getWords:(letter,from,to)=>{},
    join:(code,name)=>{},
    createSSEConnection:(code,messageHandler,notifyHandler)=>{},
    sendVote:(code,vote)=>{},
    readRules:(code)=>{},
    readScore:(code)=>{},
    getRandomPage:(nbByPage)=>{},
    startGame:(code)=>{},
    canJoin:(code)=>{},
    getNbByLetter:letter=>{},
});