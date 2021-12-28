import { x } from '@xstyled/styled-components'
import React, { FC } from 'react'
import { SF } from '../../styles/fragments'

const normalHeight = { md: '500px', xs: '300px' }
const bigHeight = { md: '900px', xs: '600px' }

const PreInfo: FC<{
    title: string
    info: string
    height?: 'normal'|'big' 
}> = ({
    title,
    info,
    height = `normal`
}) => {
    return <x.div 
        {...SF.flex}
        flexDirection="column"
        spaceY={2}
    >
        <x.h3
            fontSize={{ md: "3xl", xs: "1xl" }}
            color="gray-400"
        >
            {title}  
        </x.h3>
        <x.article
            bg={"gray-700"}
            padding={4}
            borderRadius="lg"
        >
            <x.pre
                color="gray-400"
                overflowY="auto"
                maxHeight={height === `normal` ? normalHeight : bigHeight}
                height={height === `normal` ? normalHeight : bigHeight}
                tabIndex="-1"
                style={SF.wrapLines}
            >
                {info}
            </x.pre>
        </x.article>
    </x.div>
}

export default PreInfo