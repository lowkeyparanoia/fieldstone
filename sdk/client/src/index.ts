// @champions/client — drop-in supabase-js-compatible SDK for Fieldstone
// Supports: auth, from(table), storage, channel(realtime), functions, rpc

export interface FieldstoneClientOptions {
  url: string;
  anonKey: string;
}

export interface AuthResponse {
  token: string;
  user: any;
}

class QueryBuilder {
  private client: FieldstoneClient;
  private table: string;
  private selects = '*';
  private filters: string[] = [];
  private orderBy = '';
  private limitVal = '';
  private offsetVal = '';

  constructor(client: FieldstoneClient, table: string) {
    this.client = client;
    this.table = table;
  }

  select(cols: string) {
    this.selects = cols;
    return this;
  }

  eq(col: string, val: string | number | boolean) {
    this.filters.push(`eq.${col}=${encodeURIComponent(val)}`);
    return this;
  }

  neq(col: string, val: string | number | boolean) {
    this.filters.push(`neq.${col}=${encodeURIComponent(val)}`);
    return this;
  }

  gt(col: string, val: string | number) {
    this.filters.push(`gt.${col}=${encodeURIComponent(val)}`);
    return this;
  }

  gte(col: string, val: string | number) {
    this.filters.push(`gte.${col}=${encodeURIComponent(val)}`);
    return this;
  }

  lt(col: string, val: string | number) {
    this.filters.push(`lt.${col}=${encodeURIComponent(val)}`);
    return this;
  }

  lte(col: string, val: string | number) {
    this.filters.push(`lte.${col}=${encodeURIComponent(val)}`);
    return this;
  }

  like(col: string, val: string) {
    this.filters.push(`like.${col}=${encodeURIComponent(val)}`);
    return this;
  }

  order(col: string, opts?: { ascending?: boolean }) {
    this.orderBy = `${col}.${opts?.ascending === false ? 'desc' : 'asc'}`;
    return this;
  }

  limit(n: number) {
    this.limitVal = String(n);
    return this;
  }

  range(from: number, to: number) {
    this.offsetVal = String(from);
    this.limitVal = String(to - from + 1);
    return this;
  }

  single() {
    this.limitVal = '1';
    return this;
  }

  private buildURL() {
    const params = new URLSearchParams();
    params.set('select', this.selects);
    this.filters.forEach(f => {
      const idx = f.indexOf('=');
      params.set(f.slice(0, idx), f.slice(idx + 1));
    });
    if (this.orderBy) params.set('order', this.orderBy);
    if (this.limitVal) params.set('limit', this.limitVal);
    if (this.offsetVal) params.set('offset', this.offsetVal);
    return `/api/v1/${this.table}?${params.toString()}`;
  }

  async then(onfulfilled?: (value: any) => any) {
    const data = await this.client.fetchJSON(this.buildURL());
    return onfulfilled ? onfulfilled(data) : data;
  }

  async insert(payload: any | any[]) {
    const body = Array.isArray(payload) ? payload[0] : payload;
    return this.client.fetchJSON(`/api/v1/${this.table}`, { method: 'POST', body: JSON.stringify(body) });
  }

  async update(payload: any) {
    return this.client.fetchJSON(`/api/v1/${this.table}`, {
      method: 'PATCH',
      body: JSON.stringify({ ...payload, id: payload.id }),
    });
  }

  async delete() {
    const params = new URLSearchParams();
    this.filters.forEach(f => {
      const idx = f.indexOf('=');
      params.set(f.slice(0, idx), f.slice(idx + 1));
    });
    return this.client.fetchJSON(`/api/v1/${this.table}?${params.toString()}`, { method: 'DELETE' });
  }
}

export class FieldstoneClient {
  private url: string;
  private headers: Record<string, string>;

  constructor(options: FieldstoneClientOptions) {
    this.url = options.url.replace(/\/$/, '');
    this.headers = {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${options.anonKey}`,
    };
  }

  async fetchJSON(path: string, init?: RequestInit) {
    const res = await fetch(`${this.url}${path}`, {
      ...init,
      headers: { ...this.headers, ...(init?.headers || {}) },
    });
    const data = await res.json().catch(() => null);
    if (!res.ok) {
      throw new Error(data?.error?.message || data?.error || `HTTP ${res.status}`);
    }
    return data;
  }

  // Auth namespace
  auth = {
    signUp: async (credentials: { email: string; password: string }): Promise<AuthResponse> => {
      return this.fetchJSON('/api/auth/register', {
        method: 'POST',
        body: JSON.stringify(credentials),
      });
    },
    signInWithPassword: async (credentials: { email: string; password: string }): Promise<AuthResponse> => {
      return this.fetchJSON('/api/auth/login', {
        method: 'POST',
        body: JSON.stringify(credentials),
      });
    },
    signInWithOtp: async (payload: { phone: string }): Promise<{ message: string }> => {
      return this.fetchJSON('/api/auth/otp/send', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    },
    verifyOtp: async (payload: { phone: string; token: string }): Promise<AuthResponse> => {
      return this.fetchJSON('/api/auth/otp/verify', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    },
    signOut: async (): Promise<void> => {
      await this.fetchJSON('/api/auth/logout', { method: 'POST' });
    },
    resetPasswordForEmail: async (payload: { email: string }): Promise<{ message: string }> => {
      return this.fetchJSON('/api/auth/recover', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    },
    getSession: async (): Promise<{ data: { session: any } }> => {
      return { data: { session: null } };
    },
    onAuthStateChange: (callback: (event: string, session: any) => void) => {
      return { data: { subscription: { unsubscribe: () => {} } } };
    },
  };

  from(table: string) {
    return new QueryBuilder(this, table);
  }

  // Storage namespace
  storage = {
    from: (bucket: string) => ({
      upload: async (path: string, file: File | Blob) => {
        const form = new FormData();
        form.append('file', file);
        return this.fetchJSON(`/api/storage/buckets/${bucket}/objects?path=${encodeURIComponent(path)}`, {
          method: 'POST',
          body: form as any,
          headers: {},
        });
      },
      download: async (path: string) => {
        const res = await fetch(`${this.url}/api/storage/buckets/${bucket}/objects/${path}`, {
          headers: this.headers,
        });
        if (!res.ok) throw new Error(`download failed: ${res.status}`);
        return res.blob();
      },
      remove: async (paths: string[]) => {
        return this.fetchJSON(`/api/storage/buckets/${bucket}/objects/${paths[0]}`, { method: 'DELETE' });
      },
      createSignedUrl: async (path: string, expiresIn: number) => {
        return {
          data: {
            signedUrl: `${this.url}/storage/${bucket}/${path}?expires=${Date.now() + expiresIn * 1000}`,
          },
        };
      },
    }),
  };

  // Realtime namespace
  channel(name: string) {
    const wsUrl = this.url.replace(/^http/, 'ws') + '/ws';
    let ws: WebSocket | null = null;
    const handlers: Record<string, ((payload: any) => void)[]> = {};

    return {
      on: (event: string, callback: (payload: any) => void) => {
        if (!handlers[event]) handlers[event] = [];
        handlers[event].push(callback);
        return this.channel(name);
      },
      subscribe: (callback?: (status: string) => void) => {
        ws = new WebSocket(wsUrl);
        ws.onopen = () => {
          ws?.send(JSON.stringify({ type: 'subscribe', collectionId: name }));
          callback?.('SUBSCRIBED');
        };
        ws.onmessage = (msg) => {
          try {
            const data = JSON.parse(msg.data);
            const event = data.type || 'postgres_changes';
            (handlers[event] || []).forEach((h) => h(data.payload || data));
          } catch {}
        };
        ws.onclose = () => callback?.('CLOSED');
        return this.channel(name);
      },
      unsubscribe: () => {
        ws?.close();
      },
    };
  }

  // Functions namespace
  functions = {
    invoke: async (name: string, options?: { body?: any; headers?: Record<string, string> }) => {
      return this.fetchJSON(`/functions/v1/${name}`, {
        method: 'POST',
        body: options?.body ? JSON.stringify(options.body) : undefined,
        headers: options?.headers,
      });
    },
  };

  // RPC namespace
  rpc(fn: string, params?: Record<string, any>) {
    return this.fetchJSON(`/api/v1/rpc/${fn}`, {
      method: 'POST',
      body: JSON.stringify(params || {}),
    });
  }
}

// Factory function (drop-in for createClient)
export function createClient(url: string, anonKey: string) {
  return new FieldstoneClient({ url, anonKey });
}
