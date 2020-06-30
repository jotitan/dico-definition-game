import React, {useCallback, useContext, useEffect, useState} from 'react';
import {useHistory, useParams} from 'react-router-dom'

import 'moment/locale/fr';
import GameContext from "../../context/GameContext";
import {Button, Col, Input, notification, Radio, Row, Tooltip} from 'antd';
import {DisconnectOutlined, CopyOutlined,EditOutlined, LikeOutlined, ReadOutlined} from "@ant-design/icons";
import 'antd/dist/antd.css';

import {useLocalStorage} from "../../services/local-storage.hook";
import Dico from "../dico";
import CountdownGame from "../countdown";
import Rules from "../rules";
import LikeTwoTone from "@ant-design/icons/es/icons/LikeTwoTone";
import StarTwoTone from "@ant-design/icons/es/icons/StarTwoTone";
import CheckCircleTwoTone from "@ant-design/icons/es/icons/CheckCircleTwoTone";

const TYPE_GAME_NORMAL = 'normal';

export default function Game() {
    const [name,setName] = useState('');
    const [definition,setDefinition] = useState('');
    const [vote,setVote] = useState('');
    const [context,setContext] = useState({});
    const {isGameExist,sendWordForGame,sendDefinitionForWord,join,createSSEConnection,startGame,readRules,readScore,sendVote} = useContext(GameContext);
    const [currentGame,setCurrentGame] = useLocalStorage('currentGame');
    const {TextArea} = Input;

    const {code} = useParams();
    const history = useHistory();

    const createConnection = useCallback(() => {
        createSSEConnection(code,manageEvent,manageEventNotify);
    },[createSSEConnection,code]);

    useEffect(()=>{
        isGameExist(code)
            .then(()=>{
                if(currentGame != null && currentGame.id != null && currentGame.code === code){
                    //try to reconnect
                    createConnection();
                }else{
                    // Remove cookie
                    document.cookie = 'player=; expires=Thu, 01 Jan 1970 00:00:01 GMT;path=/';
                }
            })
            .catch(e=>{
                // Notification & redirection
                notification["error"]({message:'Unknown game',description:'This game does not exist'});
                history.push('/')
            })},[isGameExist,history,code,currentGame,createConnection]);

    // Check if game exist
    const joinGame = ()=>{
        // Ask server
        join(code,name)
            .then(resp=>{
                setCurrentGame(g=>{
                    let copy = {...g};
                    copy.name = name;
                    copy.type = resp.data.type || TYPE_GAME_NORMAL;
                    // Check if current id is to different to new. If true, remove creator status and put code (case of direct access)
                    if(copy.code !== code){
                        copy.isCreator = false;
                        copy.code = code;
                    }
                    copy.id = resp.data.id;
                    return copy;
                });
            })
            .catch(err=>{
                // Game is maybe already launch
                console.log(err.response)
            })
    };

    const sendChosenWord = word=> {
        if(word ===""){return;}
        sendWordForGame(code,word);
    };

    const sendDefinition = () =>{
        sendDefinitionForWord(code,definition)
            .then(r=>console.log("Success definition"))
    };

    const manageEvent = data => {
        setContext(ctx=> {
            let newContext = {event: data.type, players: ctx.players,totalScore:ctx.totalScore,answers:ctx.answers,countdown:0,disconnect:ctx.disconnect ||[]};
            switch (data.type) {
                case "players":
                    newContext.players = data.players;
                    break;
                case "round":
                    newContext.master = data.master;
                    newContext.answers = [];
                    break;
                case "definition":
                    newContext.word = data.word;
                    newContext.master = data.master;
                    newContext.countdown = Date.now() + data.countdown*1000;
                    newContext.answers = data.answers != null ? data.answers:[];
                    break;
                case "vote":
                    newContext.definitions = data.definitions;
                    newContext.master = data.master;
                    newContext.word = data.word;
                    newContext.countdown = Date.now() + data.countdown*1000;
                    newContext.answers = data.answers != null ? data.answers:[];
                    break;
                case "rules":
                    newContext.countdown = Date.now() + data.countdown*1000;
                    newContext.answers = data.answers != null ? data.answers:[];
                    break;
                case "welcome":
                    newContext.players.push(data.player);
                    break;
                case "score":
                    newContext.roundScore = data.round;
                    newContext.totalScore = data.total;
                    newContext.detailScore = data.detail;
                    newContext.countdown = Date.now() + data.countdown*1000;
                    newContext.definitions = data.definitions != null ? data.definitions:[];
                    newContext.answer = data.answer;
                    newContext.answers = data.answers != null ? data.answers:[];
                    newContext.word = data.word;
                    break;
                default:
                    console.log("Unknown event")
            }
            window.t = newContext;
            return newContext;
        });
    };

    const manageEventNotify = data => {
        setContext(ctx=> {
            let newContext = {...ctx};
            switch (data.type) {
                case "current-score":
                    newContext.totalScore = data.total;
                    break;
                case "answer":
                    if(newContext.answers ==null){
                        newContext.answers = [];
                    }
                    newContext.answers.push(data.player);
                    newContext.countdown = Date.now() + data.countdown*1000;
                    break;
                case "disconnect":
                    newContext.disconnect.push(data.player);
                    break;
                case "reconnect":
                    let index = newContext.disconnect.indexOf(data.player);
                    if(index !== -1){
                        newContext.disconnect.splice(index,1);
                    }
                    break;
                default:console.log("Unknown event")
            }
            window.t = newContext;
            return newContext;
        });
    };

    const copyLink = ()=>{
        document.querySelector("#linkValue").select();
        document.execCommand("copy");
    };

    const showWaitingPlayers = players=>{
        return (
            <>
                <div>

                </div>
                <div className={"bandeau"}>Joueurs connectés</div>
                <div>
                    {players.map((p,i)=><div>Joueur {i+1} : {p}</div>)}
                </div>
                <div>
                    Inviter des amis
                    <input id="linkValue" value={window.location.href} size={50} style={{marginLeft:10}}/>
                    <Tooltip placement={"top"} title={"Copier le lien"}>
                        <CopyOutlined onClick={copyLink} style={{marginRight:20,fontSize:18}}/>
                    </Tooltip>
                    {currentGame.isCreator ?<Button onClick={()=>startGame(code)}>Démarrer la partie</Button>:<></>}
                </div>
            </>
        );
    };

    const showVotePanel = (definitions,master,word) => {
        if(currentGame.id === master.id && currentGame.type===TYPE_GAME_NORMAL){
            return (
                <div>
                    <div className={"bandeau"}>Vous ne pouvez pas voter, vous êtes le maitre. Voici les réponses proposées pour votre mot {word}</div>
                    {definitions.map(d=> <div style={{fontWeight:d.IsPlayerAnswer ? 'bold':'normal'}}>{d.Definition}</div>)}
                </div>
            );
        }
        return (
            <>
                {context.answers != null && context.answers.includes(currentGame.name) ? <div className={"bandeau"}>Vote enregistré</div>:
                    <div>
                        <div className={"bandeau"}>
                            Le mot est <span style={{fontWeight:'bold'}}>{word}</span>. Voter pour une des définition suivantes :
                        </div>
                        <Radio.Group onChange={value=>setVote(value.target.value)}>
                            {definitions.map((d,i)=> <Radio style={{display:'block'}} value={i} disabled={d.IsPlayerAnswer}>{d.Definition}</Radio>)}
                        </Radio.Group>
                        <Button onClick={()=>sendVote(code,vote)}>Voter</Button>
                    </div>}
            </>
        );
    };

    const showListAnswer = ()=> {
        return (
            <div>
                {context.answers != null ? context.answers.map(p=><div>{p} a répondu</div>):""}
            </div>
        )
    };

    const showGiveDefinition = (word,master) => {
        if(currentGame.id === master.id && currentGame.type===TYPE_GAME_NORMAL){
            // Can't give definition
            return (
                <>
                    <div className={"bandeau"}>Vos adversaires imaginent une définition à votre mot</div>
                    {showListAnswer()}
                </>

            )
        }else{
            return (
                <>
                    {context.answers != null && context.answers.includes(currentGame.name) ? <div className={"bandeau"}>Définition enregistrée</div>:
                        <div>
                            <div className={"bandeau"}>
                                Donnez votre définition du mot :
                                <span style={{color: 'red', fontWeight: 'bold',marginLeft:5}}>{word.replace(/&#39;/g,'\'')}</span>
                            </div>
                            < TextArea onKeyUp={value=>setDefinition(value.target.value)} rows={3} cols={80}/>
                            <Button onClick={sendDefinition}>Valider</Button>
                        </div>
                    }
                </>

            );
        }
    };

    const showNewRound = playerDico => {
        if(currentGame.id === playerDico.id){
            // Show dico to choose word
            return <Dico validateAction={sendChosenWord}/>;
        }else {
            return (
                <div>
                    <div className={"bandeau"}> <span className={"name"}>{playerDico.name}</span> est en train de choisir un mot</div>
                </div>
            );
        }
    };

    const showRules = ()=>{
        return (
            <>
                <Rules/>
                <div>
                    {context.answers != null && context.answers.includes(currentGame.name) ?
                        <div className={"bandeau"}>Vos amis lisent toujours les règles</div>:
                        <Button onClick={()=>readRules(code)}>Bien compris</Button>}

                </div>
            </>
        );
    };

    const buildView = ()=>{
        switch(context.event){
            case "players":return showWaitingPlayers(context.players);
            case "round":return showNewRound(context.master);
            case "definition":return showGiveDefinition(context.word,context.master);
            case "vote":return showVotePanel(context.definitions,context.master,context.word);
            case "score":return showScore();
            case "rules":return showRules();
            case "welcome":
                return showWaitingPlayers(context.players);
            default:
                return (
                    <Row>
                        Rejoindre la partie : <Input placeholder={"Votre nom"} onChange={v=>setName(v.target.value)}/>
                        <Button onClick={joinGame}>Go</Button>
                    </Row>);
        }
    };

    const writeN = (block,nb) => {
        let arr = Array(nb);
        for(let i = 0 ; i < nb ; i++){arr[i]=0}
        return arr.map(()=>block)
    }

    const buildIconsDetailScore = detail=> {
        //GoodDef ErrorPoint VotePoint
        return <div className={"score-icon"}>
            {detail.GoodDef ?
                <Tooltip placement="top" title={"Définition trouvée"}>
                    <CheckCircleTwoTone twoToneColor={"green"}/>
                </Tooltip>:''}
            {detail.ErrorPoint > 0?
                <Tooltip placement="top" title={"Erreur des joueurs"}>
                    {writeN(<StarTwoTone twoToneColor={"red"} />,detail.ErrorPoint)}
                </Tooltip>:''}
            {detail.VotePoint > 0 ?
                <Tooltip placement="top" title={"Vote(s) pour vous "}>
                    {writeN(<LikeTwoTone  />,detail.VotePoint)}
                </Tooltip>:''}
        </div>
    };

    const showScore = ()=> {
        if(context.totalScore == null || context.roundScore == null){return ''}
        let players = Object.keys(context.totalScore).map(p=>{return {
            name:p,
            round:context.roundScore[p],
            detail:context.detailScore[p],
            total:context.totalScore[p]
        }});
        return (
            <div style={{width:100+'%'}}>
                <div className={"bandeau"}> Résultat du tour</div>
                <div>Vraie définition du mot <span style={{fontWeight:'bold'}}>{context.word}</span> : {context.answer}</div>
                <div style={{fontWeight:'bold'}}>Les propositions</div>
                <div>
                    {Object.keys(context.definitions).map(a=><div>{a} : {context.definitions[a]}</div>)}
                </div>
                <Row className={"title-header"}>
                    <Col span={5}>Joueur</Col>
                    <Col span={8}>Manche</Col>
                    <Col span={5}>Total</Col>
                </Row>
                {players.map(p=>
                    <Row style={{fontSize:20,textAlign:'center'}} className={currentGame.name === p.name ? "score-user":""}>
                        <Col span={5}>{p.name}</Col>
                        <Col span={8}>
                            <span>+ {p.round} </span>
                            {buildIconsDetailScore(p.detail)}
                        </Col>
                        <Col span={5}>
                            <span>{p.total}</span>
                        </Col>
                    </Row>)}

                <div style={{marginTop:20,textAlign:'center'}}>
                    {context.answers != null && context.answers.includes(currentGame.name) ?
                        '':
                        <Button onClick={()=>readScore(code)}>La suite !</Button>}

                </div>
            </div>
        )
    };

    const showAnswerIcon = name => {
        if(context.answers != null && context.answers.includes(name)){
            switch(context.event){
                case "definition":return <EditOutlined style={{lineHeight:1.9,marginLeft:5}}/>;
                case "vote":return <LikeOutlined style={{lineHeight:1.9,marginLeft:5}}/>;
                default:return <></>;
            }
        }
        return <></>;
    };

    const getStatus = status => {
        switch(status){
            case "round":return "Nouveau tour";
            case "definition":return "Définition";
            case "vote":return "Votes";
            case "score":return "Score";
            case "rules":return "Règles";
            case "players":return "Attente joueur";
            default:return status;
        }
    };

    const showStatus = status => {
        return <div style={{marginTop:15}}>
            <h3>Status</h3>
            <Row className={"status"}>{status}</Row>
        </div>;
    };

    const showPanel = ()=>{
        if(context.totalScore == null){return showStatus('En attente de démarrage')}

        return (
            <>
                <div>
                    <h3>Mode de jeu</h3>
                    <Row>{currentGame.type}</Row>
                </div>
                <div>
                    <h3>Joueurs</h3>
                    {Object.keys(context.totalScore).map(key=>
                        <Row>
                            {context.disconnect.includes(key)?<DisconnectOutlined title="Déconnecté" className="icon"/>:''}
                            {context.master != null && context.master.name ===key ? <ReadOutlined title="Maître du jeu" className="icon"/>:''}
                            <span style={{fontWeight:key === currentGame.name ? "bold":"normal",marginRight:10}}>{key}</span>
                            {context.totalScore[key]} {showAnswerIcon(key)}
                        </Row>)}

                    {showStatus(getStatus(context.event))}
                    {context.countdown != null && context.countdown > new Date() ?
                        <div style={{marginTop:15}}>
                            <h3>Temps restant</h3>
                            <Row><CountdownGame text={"Temps écoulé"} time={context.countdown}/></Row>
                        </div>:''}
                </div>
            </>
        )
    };

    return (
        <Row style={{backgroundColor:'#f0f2f5'}}>
            <Col flex="200px" style={{marginLeft:20}}>
                {showPanel()}
            </Col>
            <Col flex="auto">
                {buildView()}
            </Col>
        </Row>
    )
}