"use client";

import React, { useState, useEffect } from "react";
import {
  Send,
  Plus,
  Upload,
  AlertOctagon,
  RefreshCw,
  Mail,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Zap,
  Eye,
} from "lucide-react";
import { Campaign, getCampaigns, subscribeToCampaignSSE, ProgressEvent } from "@/lib/api";
import UploadCSVModal from "@/components/UploadCSVModal";
import DLQInspectorModal from "@/components/DLQInspectorModal";
import CreateCampaignModal from "@/components/CreateCampaignModal";
import CampaignDetailsModal from "@/components/CampaignDetailsModal";

export default function Dashboard() {
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeUploadCampaign, setActiveUploadCampaign] = useState<{ id: number; title: string } | null>(null);
  const [activeInspectCampaignId, setActiveInspectCampaignId] = useState<number | null>(null);
  const [isDLQOpen, setIsDLQOpen] = useState(false);
  const [isCreateOpen, setIsCreateOpen] = useState(false);

  const fetchCampaigns = async () => {
    setLoading(true);
    try {
      const data = await getCampaigns();
      setCampaigns(data);
    } catch (err) {
      console.error("Failed to load campaigns:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCampaigns();
  }, []);

  // Listen to SSE updates for processing campaigns
  useEffect(() => {
    const unsubscribes: Array<() => void> = [];

    campaigns.forEach((c) => {
      if (c.status === "queued" || c.status === "processing") {
        const unsub = subscribeToCampaignSSE(c.id, (event: ProgressEvent) => {
          setCampaigns((prev) =>
            prev.map((item) =>
              item.id === event.campaign_id
                ? {
                  ...item,
                  sent_count: event.sent_count,
                  failed_count: event.failed_count,
                  status: event.status as any,
                }
                : item
            )
          );
        });
        unsubscribes.push(unsub);
      }
    });

    return () => {
      unsubscribes.forEach((unsub) => unsub());
    };
  }, [campaigns.map((c) => c.status).join(",")]);

  const totalSent = campaigns.reduce((acc, c) => acc + c.sent_count, 0);
  const totalFailed = campaigns.reduce((acc, c) => acc + c.failed_count, 0);
  const totalRecipients = campaigns.reduce((acc, c) => acc + c.total_recipients, 0);

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 font-sans antialiased selection:bg-indigo-500 selection:text-white">
      {/* Top Navbar */}
      <header className="sticky top-0 z-40 border-b border-zinc-800/80 bg-zinc-950/80 backdrop-blur-xl px-6 py-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-tr from-indigo-600 to-violet-500 shadow-lg shadow-indigo-500/20">
              <Send className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="text-lg font-bold tracking-tight text-white">Email Dispatcher</h1>
              <p className="text-[11px] text-zinc-400 font-mono">Monorepo Production Engine v1.0</p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={() => setIsDLQOpen(true)}
              className="flex items-center gap-2 rounded-xl bg-red-500/10 border border-red-500/20 px-4 py-2 text-xs font-semibold text-red-400 hover:bg-red-500/20 transition-all"
            >
              <AlertOctagon className="w-4 h-4" />
              <span>DLQ Inspector</span>
            </button>

            <button
              onClick={() => setIsCreateOpen(true)}
              className="flex items-center gap-2 rounded-xl bg-indigo-600 px-4 py-2 text-xs font-semibold text-white hover:bg-indigo-500 transition-all shadow-lg shadow-indigo-600/30"
            >
              <Plus className="w-4 h-4" />
              <span>New Campaign</span>
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-8">
        {/* Metrics Grid */}
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4 mb-8">
          <div className="rounded-2xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm">
            <div className="flex items-center justify-between text-zinc-400 mb-2">
              <span className="text-xs font-medium uppercase tracking-wider">Total Campaigns</span>
              <Mail className="w-4 h-4 text-indigo-400" />
            </div>
            <p className="text-2xl font-bold text-white">{campaigns.length}</p>
          </div>

          <div className="rounded-2xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm">
            <div className="flex items-center justify-between text-zinc-400 mb-2">
              <span className="text-xs font-medium uppercase tracking-wider">Emails Dispatched</span>
              <CheckCircle2 className="w-4 h-4 text-emerald-400" />
            </div>
            <p className="text-2xl font-bold text-white">{totalSent.toLocaleString()}</p>
          </div>

          <div className="rounded-2xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm">
            <div className="flex items-center justify-between text-zinc-400 mb-2">
              <span className="text-xs font-medium uppercase tracking-wider">Failed Deliveries</span>
              <AlertTriangle className="w-4 h-4 text-red-400" />
            </div>
            <p className="text-2xl font-bold text-white">{totalFailed.toLocaleString()}</p>
          </div>

          <div className="rounded-2xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm">
            <div className="flex items-center justify-between text-zinc-400 mb-2">
              <span className="text-xs font-medium uppercase tracking-wider">Total Target Queue</span>
              <Zap className="w-4 h-4 text-amber-400" />
            </div>
            <p className="text-2xl font-bold text-white">{totalRecipients.toLocaleString()}</p>
          </div>
        </div>

        {/* Campaign List Header */}
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-bold text-white">Active Campaigns</h2>
          <button
            onClick={fetchCampaigns}
            className="flex items-center gap-1.5 rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-1.5 text-xs text-zinc-400 hover:text-white transition-colors"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin" : ""}`} />
            <span>Refresh</span>
          </button>
        </div>

        {/* Campaign Cards */}
        {campaigns.length === 0 ? (
          <div className="rounded-2xl border border-zinc-800/80 bg-zinc-900/30 p-12 text-center">
            <Mail className="w-12 h-12 text-zinc-600 mx-auto mb-3" />
            <h3 className="text-base font-semibold text-zinc-300">No Campaigns Found</h3>
            <p className="text-xs text-zinc-500 mt-1 mb-4">Get started by creating your first email campaign.</p>
            <button
              onClick={() => setIsCreateOpen(true)}
              className="inline-flex items-center gap-2 rounded-xl bg-indigo-600 px-4 py-2 text-xs font-semibold text-white hover:bg-indigo-500 transition-all"
            >
              <Plus className="w-4 h-4" />
              <span>Create Campaign</span>
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {campaigns.map((c) => {
              const progressPct = c.total_recipients > 0
                ? Math.min(100, Math.round(((c.sent_count + c.failed_count) / c.total_recipients) * 100))
                : 0;

              return (
                <div
                  key={c.id}
                  onClick={() => setActiveInspectCampaignId(c.id)}
                  className="rounded-2xl border border-zinc-800 bg-zinc-900/60 p-6 flex flex-col justify-between hover:border-indigo-500/40 hover:bg-zinc-900/80 transition-all shadow-xl cursor-pointer group"
                >
                  <div>
                    <div className="flex items-start justify-between mb-3">
                      <h3 className="font-bold text-white text-base truncate max-w-[200px] group-hover:text-indigo-300 transition-colors">
                        {c.title}
                      </h3>
                      <span
                        className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wider ${
                          c.status === "completed"
                            ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                            : c.status === "queued" || c.status === "processing"
                              ? "bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 animate-pulse"
                              : "bg-zinc-800 text-zinc-400"
                        }`}
                      >
                        {c.status}
                      </span>
                    </div>

                    <p className="text-xs text-zinc-400 mb-4 line-clamp-1">
                      <span className="text-zinc-500">Subject:</span> {c.subject}
                    </p>

                    {/* Progress Bar */}
                    <div className="space-y-1.5 mb-4">
                      <div className="flex justify-between text-xs font-medium">
                        <span className="text-zinc-400">Dispatch Progress</span>
                        <span className="text-indigo-400">{progressPct}%</span>
                      </div>
                      <div className="h-2 w-full rounded-full bg-zinc-800 overflow-hidden">
                        <div
                          className="h-full bg-gradient-to-r from-indigo-500 to-emerald-400 transition-all duration-500"
                          style={{ width: `${progressPct}%` }}
                        />
                      </div>
                      <div className="flex justify-between text-[11px] text-zinc-500">
                        <span>Sent: {c.sent_count}</span>
                        <span>Failed: {c.failed_count}</span>
                        <span>Total: {c.total_recipients}</span>
                      </div>
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="pt-4 border-t border-zinc-800/80 flex items-center justify-between">
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        setActiveInspectCampaignId(c.id);
                      }}
                      className="flex items-center gap-1 text-[11px] font-semibold text-indigo-400 hover:text-indigo-300 transition-colors"
                    >
                      <Eye className="w-3.5 h-3.5" />
                      <span>View Details</span>
                    </button>

                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        setActiveUploadCampaign({ id: c.id, title: c.title });
                      }}
                      className="flex items-center gap-1.5 rounded-lg bg-zinc-800 px-3 py-1.5 text-xs font-medium text-zinc-200 hover:bg-zinc-700 hover:text-white transition-colors"
                    >
                      <Upload className="w-3.5 h-3.5 text-indigo-400" />
                      <span>Upload CSV</span>
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </main>

      {/* Modals */}
      {activeInspectCampaignId && (
        <CampaignDetailsModal
          campaignId={activeInspectCampaignId}
          isOpen={!!activeInspectCampaignId}
          onClose={() => setActiveInspectCampaignId(null)}
        />
      )}

      {activeUploadCampaign && (
        <UploadCSVModal
          campaignId={activeUploadCampaign.id}
          campaignTitle={activeUploadCampaign.title}
          isOpen={!!activeUploadCampaign}
          onClose={() => setActiveUploadCampaign(null)}
          onSuccess={fetchCampaigns}
        />
      )}

      <DLQInspectorModal
        isOpen={isDLQOpen}
        onClose={() => setIsDLQOpen(false)}
      />

      <CreateCampaignModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        onSuccess={fetchCampaigns}
      />
    </div>
  );
}
