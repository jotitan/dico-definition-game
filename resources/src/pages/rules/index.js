import React from 'react';

export default function Rules() {

    return ( <div>
            <h2>Règles du jeu</h2>
            Chaque tour se compose des étapes suivantes :
            <ul>
                <li>Le maître du jeu choisit un mot dans le dictionnaire</li>
                <li>Chaque joueur donne une définition qu'il suppose plausible pour le mot</li>
                <li>Chaque joueur vote pour la définition qu'il lui semble la plus réaliste</li>
                <li>On change de maître du jeu</li>
            </ul>

            Les poits sont calculés de la manière suivate :
            <ul>
                <li>Un point pour chaque joueur qui a voté pour la bonne réponse</li>
                <li>Un point pour le maître du jeu quand quelqu'un a voté pour à la bonne définition</li>
                <li>Deux points pour le joueur dont la définition a été vôté</li>
            </ul>

        </div>
    )
}