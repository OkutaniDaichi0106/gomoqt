export type Version = number;

export const Versions = {
  LITE_DRAFT_01: 0xff0dad01 as Version,
  LITE_DRAFT_02: 0xff0dad02 as Version,
} as const;

export const DEFAULT_VERSION: Version = Versions.LITE_DRAFT_01;

export const DEFAULT_CLIENT_VERSIONS: Set<Version> = new Set([
  Versions.LITE_DRAFT_01,
]);
