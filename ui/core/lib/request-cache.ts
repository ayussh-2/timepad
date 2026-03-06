import dayjs from "dayjs";

// TTLs in ms
const TODAY_TTL = 60_000; // 1 min – today's data changes frequently
const PAST_TTL = 10 * 60_000; // 10 min – historical data is stable

interface Entry<T> {
    data: T;
    expiry: number;
}

const store = new Map<string, Entry<unknown>>();
const pending = new Map<string, Promise<unknown>>();

/**
 * Wraps a fetcher with:
 *   - In-flight deduplication: concurrent callers for the same key share one request.
 *   - TTL caching: resolved data is served from memory until expiry.
 */
export function cached<T>(
    key: string,
    fetcher: () => Promise<T>,
    ttl: number,
): Promise<T> {
    const hit = store.get(key) as Entry<T> | undefined;
    if (hit && hit.expiry > Date.now()) return Promise.resolve(hit.data);

    const inflight = pending.get(key) as Promise<T> | undefined;
    if (inflight) return inflight;

    const p = fetcher()
        .then((data) => {
            store.set(key, { data, expiry: Date.now() + ttl });
            pending.delete(key);
            return data;
        })
        .catch((err) => {
            pending.delete(key);
            throw err;
        });

    pending.set(key, p);
    return p;
}

/** Returns true if cache has a non-expired entry for `key`. */
export function hasFresh(key: string): boolean {
    const hit = store.get(key);
    return !!hit && hit.expiry > Date.now();
}

/** Remove all cache entries whose key starts with `prefix`. */
export function bust(prefix: string): void {
    for (const k of store.keys()) {
        if (k.startsWith(prefix)) store.delete(k);
    }
}

/** Returns the appropriate TTL for a given date string (YYYY-MM-DD). */
export function ttlFor(date: string): number {
    return date === dayjs().format("YYYY-MM-DD") ? TODAY_TTL : PAST_TTL;
}
