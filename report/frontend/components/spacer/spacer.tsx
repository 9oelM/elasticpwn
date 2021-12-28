import { x } from '@xstyled/styled-components'
import React, { FC } from 'react'

const Spacer: FC<{
    vertical?: boolean
    size: string
}> = ({
    vertical = true,
    size
}) => {
    return <x.div {...(vertical ? {
        mt: size
    } : {
        mr: size,
    })}></x.div>
}

export default Spacer