export type NetworkStatus = { network: unknown; net_info?: unknown };
export type UnsignedTransaction = {
  type: "ogo.unsigned_transaction.v1";
  development_only: boolean;
  chain_id: string;
  from: string;
  messages: unknown[];
  fee?: unknown;
  memo?: string;
  account_number: number;
  sequence: number;
  signing: { client_side_only: true; mode: "direct"; private_keys_requested: false };
};

export interface SignedTransaction {
  /** Base64-encoded bytes produced by the client-side wallet. */
  signed_tx: string;
  /** Stable application/payment id used for retry-safe broadcast. */
  idempotencyKey: string;
}

export class OgoNetworkClient {
  constructor(private readonly baseUrl: string) {}

  private async request<T>(path: string, init?: RequestInit): Promise<T> {
    const response = await fetch(`${this.baseUrl.replace(/\/$/, "")}${path}`, {
      ...init,
      headers: { "content-type": "application/json", ...(init?.headers || {}) },
    });
    const body = await response.json();
    if (!response.ok) throw new Error(body?.error?.message || `OGO API error ${response.status}`);
    return body as T;
  }

  getNetwork(): Promise<NetworkStatus> { return this.request("/v1/network"); }
  getBalance(address: string): Promise<unknown> { return this.request(`/v1/balances/${encodeURIComponent(address)}`); }
  getTransaction(hash: string): Promise<unknown> { return this.request(`/v1/transactions/${encodeURIComponent(hash)}`); }
  getAddressTransactions(address: string): Promise<unknown> { return this.request(`/v1/addresses/${encodeURIComponent(address)}/transactions`); }
  getValidators(): Promise<unknown> { return this.request("/v1/validators"); }
  getEpochs(): Promise<unknown> { return this.request("/v1/epochs"); }
  prepareTransaction(input: Record<string, unknown>): Promise<UnsignedTransaction> {
    return this.request("/v1/transactions/prepare", { method: "POST", body: JSON.stringify(input) });
  }
  broadcastSignedTransaction(input: SignedTransaction): Promise<unknown> {
    return this.request("/v1/transactions/broadcast", {
      method: "POST",
      headers: { "idempotency-key": input.idempotencyKey },
      body: JSON.stringify({ signed_tx: input.signed_tx }),
    });
  }
  verifyTransaction(hash: string): Promise<unknown> {
    return this.getTransaction(hash).then((tx: any) => ({ committed: tx?.result?.tx_result?.code === 0, transaction: tx }));
  }
}
