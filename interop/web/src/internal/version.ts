export type Version = bigint;

export const Versions = {
    DEVELOP: 0xffffff00n as Version,
} as const;

export const DEFAULT_VERSION: Version = Versions.DEVELOP;