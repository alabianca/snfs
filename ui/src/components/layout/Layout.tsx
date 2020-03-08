import React from 'react';
import './Layout.css';

const Layout = (props: any) => (
    <div className="layout">
        {props.children}
        {/* <div className="sidebar">{props.sidebar}</div>
        <div className="content">{props.content}</div> */}
    </div>
)

export default Layout;