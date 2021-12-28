import { x } from '@xstyled/styled-components'
import React, { FC } from 'react'
import { SF } from '../../styles/fragments'
import { ElasticProductInfo } from '../../types/elastic'

const TableInfo: FC<{
    title: string,
    headings: string[]
}> = ({
    children,
    title,
    headings
}) => {
        return <x.div
            {...SF.flex}
            flexDirection="column"
            spaceY={2}
            maxHeight={{ md: '500px', xs: '300px' }}
        >
            <x.h3
                fontSize={{ md: "3xl", xs: "1xl" }}
                color="gray-400"
            >
                {title}
            </x.h3>
            <x.div
                height="100%"
                overflowY="auto"
                borderRadius="lg"
            >
            <x.table
                tableLayout="auto"
                color="gray-400"
                bg="gray-700"
                width="100%"
                
                borderRadius="lg"
                padding={4}
            >
                <x.thead
                    borderWidth={1}
                >
                    <x.tr>
                        {headings.map((heading, i) => {
                            return <x.th key={i}>{heading}</x.th>
                        })}
                    </x.tr>
                </x.thead>
                <x.tbody
                >
                    {children}
                </x.tbody>
            </x.table>
            </x.div>
        </x.div>
    }

export default TableInfo