import React from 'react'
import { x } from "@xstyled/styled-components";
import { enhance } from "../../util/react-essentials";
import { SF } from '../../styles/fragments';
import { usePersistentDatabaseServerAvailableStatus } from '../../hooks/usePersistentDatabaseServerAvailableStatus';

export type QuickActionButtonsProps = {
    onReviewed: VoidFunction 
    onPass: VoidFunction
}

export const QuickActionButtons = enhance<QuickActionButtonsProps>(({
    onReviewed,
    onPass
}) => {
    return <x.section
        w={1}
        position='fixed'
        bottom={0}
        mb={6}
        spaceX={2}
        {...SF.flex}
        justifyContent='center'
        pointerEvents="none"
    >
        <x.button 
            {...SF.standardButton}
            onClick={onReviewed}
        >
            Reviewed
            <x.p pt={2}>Alt+R</x.p>
        </x.button>
        <x.button 
            {...SF.standardButton}
            onClick={onPass}
        >
            Come back later
            <x.p pt={2}>Alt+C</x.p>
        </x.button>
    </x.section>
})()