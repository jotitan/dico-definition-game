import React, {useContext, useState} from 'react';
import 'moment/locale/fr';
import {Button, Input, notification, Row} from 'antd';
import GameContext from "../../context/GameContext";
import {useLocalStorage} from '../../services/local-storage.hook'

export default function Home() {
    const [codeGame,setCodeGame] = useState('');
    const {createGame,createCurrentGame,canJoin} = useContext(GameContext);
    const [currentGame,setCurrentGame] = useLocalStorage('currentGame');
    const doCreateGame = ()=>{
        // Ask server
        createGame().then(resp=>launchGame(resp.data.code))
            .catch(err=>console.log(err))
    };

    const launchGame = (code,isCreator=true) => {
        if(code === ""){
            notification["error"]({message:'Empty code',description:'Code must be define'});
            return;
        }
        if(currentGame != null && currentGame.code !== code){
            setCurrentGame(null);
        }
        // Check if game is already started
        canJoin(code).then(res=>{
            if(res.data.playable) {
                // Save game in context, define user as creator and admin
                setCurrentGame(createCurrentGame(code, isCreator));
                let basename = "";
                if(window.location.href.indexOf('/dico_game')!==-1){
                    // Production case, modify basename in router
                    basename = "/dico_game"
                }
                window.location.href = `${basename}/game/${code}`;
            }else{
                notification["error"]({message:'Game started',description:'Game is already started, impossible to join'});
            }
        })
    };

    const gotoGame = ()=>launchGame(codeGame,false);

    return (
        <>
            <div className="App">
                <Row>
                    <Button onClick={doCreateGame}>Créer une partie</Button>
                </Row>
                <Row>Ou</Row>
                <Row>
                    Rejoindre une partie : <Input placeholder={"Code"} onChange={v=>setCodeGame(v.target.value)} style={{width:60}}/>
                    <Button onClick={gotoGame}>Jouer !</Button>
                </Row>
            </div>
        </>
    )
}