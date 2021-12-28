export const SF = {
    fullWH: {
        w: `100%`,
        h: `100%`,
    },
    flex: {
        display: 'flex',
    },
    wrapLines: {
        // xstyled does not support this
        wordWrap: "break-word",
        'whiteSpace': "pre-wrap"
    },
    cursorPointer: {
        cursor: 'pointer'
    },
    standardButton: {
        w: 100,
        h: 100,
        bg: { hover: 'gray-500', _: 'gray-600' }, 
        borderRadius: 6,
        display: `flex`,
        flexDirection: 'column',
        justifyContent: 'center',
        alignItems: 'center',
        pointerEvents: "auto",
        transition: true,
        transitionDuration: 200,
        boxShadow: "2xl",
    }
}