import React from 'react';
import 'moment/locale/fr';
import Countdown from "react-countdown";

export default function CountdownGame({time,text}) {

    const renderer = ({ hours, minutes, seconds, completed,milliseconds }) => {
        let totalSeconds = hours * 3600 + minutes * 60 + seconds;
        let strTotal = totalSeconds.toString().padStart(2,"0")
        let strMillis = milliseconds.toString().padStart(3,"0");
        if (completed) {
            // Render a complete state
            return text;
        } else {
            return (
                totalSeconds<=5 ?
                    <span style={{color:'red',fontWeight:'bold'}}>{strTotal}.{strMillis}</span>
                :
                    <span style={{color:'black',fontWeight:'normal'}}>{strTotal}</span>
            );
        }
    };
    return (<>
        <Countdown date={time} renderer={renderer} intervalDelay={10} precision={3}/>
        </>
    )
}