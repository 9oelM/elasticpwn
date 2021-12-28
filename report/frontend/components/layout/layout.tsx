import React, { FC } from 'react'
import {
    defaultTheme,
    ThemeProvider,
    Preflight,
  } from '@xstyled/styled-components'

const Layout: FC = ({ children }) => {
    return <ThemeProvider
        theme={defaultTheme}
    >
        <Preflight/>
        {children}
    </ThemeProvider>
} 

export default Layout