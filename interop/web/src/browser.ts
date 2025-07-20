// Browser-specific entry point for the MoQT web library
// This file exports only browser-compatible functionality

export * from './session';
export * from './broadcast_path';
export * from './track_prefix';
export * from './options';
export * from './info';
export * from './client';
export * from './announce_stream';
export * from './group_stream';
export * from './session_stream';
export * from './stream_type';
export * from './subscribe_stream';
export * from './track';
export * from './track_mux';

// Browser-specific exports or polyfills can be added here
// For example, if there are Node.js-specific features that need browser alternatives
