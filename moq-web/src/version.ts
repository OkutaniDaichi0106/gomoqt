export type Version = number;

export const Versions = {
	LITE_DRAFT_01: 0xff0dad01 as Version,
	LITE_DRAFT_02: 0xff0dad02 as Version,
	DEVELOPMENT: 0xfeedbabe as Version,
} as const;

export const DEFAULT_VERSION: Version = Versions.DEVELOPMENT;

export const DEFAULT_CLIENT_VERSIONS: Set<Version> = new Set([
	DEFAULT_VERSION,
]);
