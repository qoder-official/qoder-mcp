declare module 'crypto' {
  export function randomUUID(): string;
}

declare module 'node:crypto' {
  export * from 'crypto';
} 