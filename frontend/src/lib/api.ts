function getApiBaseUrl(): string {
  if (process.env.NEXT_PUBLIC_API_URL) return process.env.NEXT_PUBLIC_API_URL;
  if (typeof window !== "undefined") {
    const host = window.location.hostname || "127.0.0.1";
    return `http://${host}:8080`;
  }
  return "http://127.0.0.1:8080";
}

export interface Campaign {
  id: number;
  title: string;
  subject: string;
  body_template: string;
  status: 'draft' | 'queued' | 'processing' | 'completed' | 'failed' | 'paused';
  total_recipients: number;
  sent_count: number;
  failed_count: number;
  created_at: string;
  updated_at: string;
}

export interface DLQRecord {
  id: number;
  job_id: string;
  campaign_id: number;
  recipient_email: string;
  error_reason: string;
  payload_json: string;
  status: 'pending' | 'replayed' | 'discarded';
  failed_at: string;
}

export interface ProgressEvent {
  campaign_id: number;
  status: string;
  total_recipients: number;
  sent_count: number;
  failed_count: number;
  progress_pct: number;
}

export interface Recipient {
  id: number;
  campaign_id: number;
  name: string;
  email: string;
  status: 'pending' | 'sent' | 'failed';
  error_message?: string;
  sent_at?: string;
  created_at: string;
}

export interface CampaignDetails extends Campaign {
  recipients: Recipient[];
}

// Fetch all campaigns
export async function getCampaigns(): Promise<Campaign[]> {
  const res = await fetch(`${getApiBaseUrl()}/api/v1/campaigns`, { cache: "no-store" });
  if (!res.ok) throw new Error("Failed to fetch campaigns");
  return res.json();
}

// Fetch single campaign details with recipients
export async function getCampaignDetails(id: number): Promise<CampaignDetails> {
  const res = await fetch(`${getApiBaseUrl()}/api/v1/campaigns/${id}`, { cache: "no-store" });
  if (!res.ok) throw new Error("Failed to fetch campaign details");
  return res.json();
}

// Create new campaign draft
export async function createCampaign(data: { title: string; subject: string; body_template: string }): Promise<Campaign> {
  const res = await fetch(`${getApiBaseUrl()}/api/v1/campaigns`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Failed to create campaign");
  return res.json();
}

// Upload CSV recipient list
export async function uploadCSV(campaignId: number, file: File) {
  const formData = new FormData();
  formData.append("campaign_id", campaignId.toString());
  formData.append("file", file);

  const res = await fetch(`${getApiBaseUrl()}/api/v1/campaigns/upload`, {
    method: "POST",
    body: formData,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Upload failed" }));
    throw new Error(err.error || "Failed to upload CSV");
  }
  return res.json();
}

// Fetch DLQ records
export async function getDLQRecords(statusFilter?: string): Promise<DLQRecord[]> {
  const url = statusFilter
    ? `${getApiBaseUrl()}/api/v1/dlq?status=${statusFilter}`
    : `${getApiBaseUrl()}/api/v1/dlq`;
  const res = await fetch(url, { cache: "no-store" });
  if (!res.ok) throw new Error("Failed to fetch DLQ records");
  return res.json();
}

// Replay DLQ failed job
export async function replayDLQRecord(id: number) {
  const res = await fetch(`${getApiBaseUrl()}/api/v1/dlq/${id}/replay`, { method: "POST" });
  if (!res.ok) throw new Error("Failed to replay DLQ job");
  return res.json();
}

// Subscribe to real-time SSE stream for a campaign
export function subscribeToCampaignSSE(
  campaignId: number,
  onData: (event: ProgressEvent) => void,
  onError?: (err: any) => void
): () => void {
  const eventSource = new EventSource(`${getApiBaseUrl()}/api/v1/campaigns/${campaignId}/stream`);

  eventSource.onmessage = (e) => {
    try {
      const data: ProgressEvent = JSON.parse(e.data);
      onData(data);
    } catch (err) {
      if (onError) onError(err);
    }
  };

  eventSource.onerror = (err) => {
    if (onError) onError(err);
    eventSource.close();
  };

  // Return unsubscribe function
  return () => eventSource.close();
}
