import React from 'react'
import './sidebar.css'
import Lib from '../../lib';
import Header from '../header/header'



export type WindowAction = "close" | "minimize" | "maximize"


class Sidebar extends React.Component {

    public handleWindowAction(windowAction: WindowAction) {
        console.log(windowAction)
        try {
            Lib.getIpcRenderer().send(`window:${windowAction}`)
        } catch(e) {
            console.log("IPC Renderer is not available")
        }
    }
    
    public render() {
        return (
            <div className="sidebar">
                <Header onMinimize={() => this.handleWindowAction('minimize')} onMaximize={() => this.handleWindowAction('maximize')} onClose={() => this.handleWindowAction('close')}/>
                <div className="level with-pad">
                    <div className="level-left">
                        <div className="level-item">
                            <span id="node-title">Nodes</span>
                        </div>
                    </div>
                    <div className="level-right">
                        <div className="level-item">
                            <button className="button is-primary is-small is-outlined"> + New</button>
                        </div>
                    </div>
                </div>
            </div>
        )
    }
}

// const Sidebar = (props: any) => (
//     <div className="sidebar">
//         <Header/>
//         <div className="level">
//             <div className="level-left">
//                 <div className="level-item">
//                     <span>Nodes</span>
//                 </div>
//             </div>
//             <div className="level-right">
//                 <div className="level-item">
//                     <button>New</button>
//                 </div>
//             </div>
//         </div>
//     </div>
// )

export default Sidebar