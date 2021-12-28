export enum SizeUnit {
    b = "b",
    kb = "kb",
    mb = "mb",
    gb = "gb",
    tb = "tb",
    pb = "pb",
    eb = "eb",
    zb = "zb",
}

export function getUnitScore(unit: SizeUnit) {
    switch (unit) {
        case SizeUnit.b:
            return 0
        case SizeUnit.kb:
            return 1
        case SizeUnit.mb:
            return 2
        case SizeUnit.gb:
            return 3
        case SizeUnit.tb:
            return 4
        case SizeUnit.pb:
            return 5
        case SizeUnit.eb:
            return 6
        case SizeUnit.zb:
            return 7
        // idk but this case happens
        default:
            return 0
    }
}