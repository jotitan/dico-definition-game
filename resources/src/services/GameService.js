import axios from "axios";
import {notification} from "antd";

function joinGame(code){
    console.log(code)
}

function getBaseUrl(){
    return window.location.href.indexOf('/dico_game') !== -1 ? '/dico_game':'';
}

function createGame(){
    return axios({
            url: getBaseUrl() + '/api/game/create',
            method:'POST'
        }
    );
}

function canJoin(code){
    return axios({
        url:getBaseUrl() + `/api/game/can_join/${code}`,
        method:'GET'
    })
}

function isGameExist(code){
    return axios({url: getBaseUrl() + `/api/game/detail/${code}`,method:'GET'});
}

function createCurrentGame(code,isCreator=false){
    return {code:code,isCreator:isCreator,name:''};
}

function sendWordForGame(code,word){
    return axios({
        url: getBaseUrl() + `/api/game/choose_word/${code}`,
        data:`{"word":"${word}"}`,
        method:'POST'
    });
}

function sendDefinitionForWord(code,definition){
    return axios({
        url: getBaseUrl() + `/api/game/give_definition/${code}`,
        data:`{"definition":"${encodeURIComponent(definition)}"}`,
        method:'POST'
    });
}

function getWords(letter,from,to){
    return axios({
        url: getBaseUrl() + `/api/dico/search/${letter}/${from}/${to}`,
        method:'GET'
    })
}

function getNbByLetter(letter){
    return axios({
        url: getBaseUrl() + `/api/dico/nb/${letter}`,
        method:'GET'
    })
}

function join(code,name){
    return axios({
            url: getBaseUrl() + `/api/game/join/${code}`,
            method:'POST',
            data:JSON.stringify({name:name})
        }
    )
}

function createSSEConnection(code,messageHandler,notifyHandler){
    let evtSrc = new EventSource(getBaseUrl() + `/api/game/connect/${code}`);
    evtSrc.addEventListener("message",event=>{
        let data = JSON.parse(event.data);
        messageHandler(data);
    });
    evtSrc.addEventListener("notify",event=>{
        let data = JSON.parse(event.data);
        notifyHandler(data);
    });
};

function startGame(code){
    axios({
        url: getBaseUrl() + `/api/game/start/${code}`,
        method:'GET'
    }).then(d=>console.log("Partie demarre"))
};

function readScore(code){
    axios({
        url: getBaseUrl() + `/api/game/read_score/${code}`,
        method:'POST'
    }).then(d=>console.log("En attente"))
};


function readRules(code){
    axios({
        url: getBaseUrl() + `/api/game/read_rules/${code}`,
        method:'POST'
    }).then(d=>console.log("En attente"))
};

function sendVote(code,vote){
    axios({
        url: getBaseUrl() + `/api/game/vote/${code}`,
        method:'POST',
        data:`{"vote":${vote}}`
    })
        .then(d=>console.log("A voté"))
        .catch(e=>notification["error"]({message:'Vote impossible',description:'Impossible de voter'}))
};


export default {
    joinGame:joinGame,
    createGame:createGame,
    createCurrentGame:createCurrentGame,
    isGameExist:isGameExist,
    getNbByLetter:getNbByLetter,
    sendWordForGame:sendWordForGame,
    sendDefinitionForWord:sendDefinitionForWord,
    getWords:getWords,
    join:join,
    sendVote:sendVote,
    readRules:readRules,
    readScore:readScore,
    startGame:startGame,
    canJoin:canJoin,
    createSSEConnection:createSSEConnection
};
