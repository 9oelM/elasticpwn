export function getOnlyString(alphaNumericWord: string): string {
    return alphaNumericWord.replace(/[^a-z]/gi, '');
}

export function getOnlyNumber(alphaNumericWord: string): number | null {
    const floatingNumberMatches = alphaNumericWord.match(/[+-]?\d+(\.\d+)?/g)
    if (floatingNumberMatches === null) return null

    return parseFloat(floatingNumberMatches[0])
}
