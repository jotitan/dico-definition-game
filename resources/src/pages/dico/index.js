import React, {useContext, useState} from 'react';
import 'moment/locale/fr';
import GameContext from "../../context/GameContext";
import {Button, Col, Row, Slider} from "antd";

export default function Dico({validateAction=(value)=>{console.log("Selected",value)}}) {
    const [words,setWords] = useState([]);
    const [selectedLetter,setSelectedLetter] = useState('');
    const {getWords,getNbByLetter} = useContext(GameContext);
    const [nbWords,setNbWords] = useState(0);
    const nbByPage = 20;
    const [selectedWord,setSelectedWord] = useState('');
    const [currentPage,setCurrentPage] = useState(0);

    const loadWords = letter=>{
        setSelectedLetter(letter);
        setCurrentPage(0);
        loadWordsPages(letter,0,nbByPage);
        getNbByLetter(letter).then(resp=>setNbWords(resp.data.nb-1))
    };
    const loadWordsPages = (letter,from,to)=>getWords(letter,from,to).then(resp=>setWords(resp.data));

    const showWord = wordDef => {
        return (
            <Row>
                <Col className={"word"+(selectedWord===wordDef.Word?" selected-word":"")} onClick={()=>setSelectedWord(wordDef.Word)}>{wordDef.Word.replace(/&#39;/g,'\'')}</Col>
                <Col>: {wordDef.Definition}</Col>
            </Row>
        )
    };

    const showWords = ()=> {
        return (
            words.map(w=>showWord(w))
        );
    };
    const showLetters = ()=>{
        let letters = [];
        for(let ascii = 65 ; ascii < 91 ; ascii++){
            letters.push(String.fromCharCode(ascii));
        }
        return (
            letters.map(l=>
                <div className={"letter" + ((selectedLetter === l) ? " selected-letter" : "")}
                     onClick={() => loadWords(l)}>{l}</div>
            )
        );
    };

    return (
        <>
            <div>
                <h3>
                    Choisissez un mot dans le dictionnaire :
                    <span style={{fontWeight:'bold',color:'red',marginLeft:5}}>{selectedWord.replace(/&#39;/g,'\'')}</span>
                    <Button onClick={()=>validateAction(selectedWord)} style={{marginLeft:10}}>Valider</Button>
                </h3>
                {showLetters()}
                <div className="letter">-</div>
                <Slider defaultValue={currentPage} value={currentPage} max={Math.ceil(nbWords/nbByPage)-1}
                        style={{width:300,display:'inline-block'}}
                        onChange={page=>{
                            setCurrentPage(page);
                            loadWordsPages(selectedLetter,page*nbByPage,(page+1)*nbByPage);
                        }}/>

            </div>
            <div style={{clear:"both"}}></div>
            <hr/>
            <div>{showWords()}</div>
        </>
    )
}