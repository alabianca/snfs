import React from 'react';
import './header.css';

interface HeaderProps {
    onClose: () => void,
    onMinimize: () => void,
    onMaximize: () => void,
}

const Header = ({onClose, onMinimize, onMaximize}: HeaderProps) => 
(<div className="level">
    <div className="level-left header-action window-actions">
        <div className="level-item">
            <div className="traffic-lights">
                <button onClick={ () => onClose() } className="traffic-light close"></button>
                <button onClick={ () => onMinimize() } className="traffic-light minimize"></button>
                <button onClick={ () => onMaximize() } className="traffic-light expand"></button>
            </div>
        </div>
    </div>
    <div className="level-right header-action"></div>
</div>)

export default Header;