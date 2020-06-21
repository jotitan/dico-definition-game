import React from 'react';
import {BrowserRouter as Router, Route, Switch} from 'react-router-dom'
import GameContext from './context/GameContext'
import GameService from './services/GameService'
import './App.css';
import {Layout} from 'antd';

import Home from "./pages/home";
import Game from "./pages/game";

function App() {
    const {Header,Content} = Layout;
    let basename = "";
    if(window.location.href.indexOf('/dico_game')!==-1){
        // Production case, modify basename in router
        basename = "/dico_game"
    }
    return (
        <Layout>
            <Header className="header">

                <a href={basename + "/"}>
                    <img src={"/logo192.png"} style={{width:50,height:50}} alt={"Logo"}/>
                    Jeu du dictionnaire
                </a>
            </Header>
            <Content  className={"content"}>
                <GameContext.Provider value={GameService}>
                    <Router basename={basename}>
                        <Switch>
                            <Route path="/game/:code" component={Game} exact/>
                            <Route path="/" component={Home} exact/>
                        </Switch>
                    </Router>
                </GameContext.Provider>
            </Content>
        </Layout>


    );
}

export default App;
