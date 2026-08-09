"use client";

import React, { useState, useEffect } from "react";
import {
  X,
  Search,
  CheckCircle2,
  AlertTriangle,
  Clock,
  RefreshCw,
  Mail,
  User,
  FileText,
  Send,
} from "lucide-react";
import { CampaignDetails, getCampaignDetails, dispatchCampaign, Recipient } from "@/lib/api";

interface CampaignDetailsModalProps {
  campaignId: number;
  isOpen: boolean;
  onClose: () => void;
}

export default function CampaignDetailsModal({
  campaignId,
  isOpen,
  onClose,
}: CampaignDetailsModalProps) {
  const [details, setDetails] = useState<CampaignDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [isDispatching, setIsDispatching] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  const fetchDetails = async () => {
    setLoading(true);
    try {
      const data = await getCampaignDetails(campaignId);
      setDetails(data);
    } catch (err) {
      console.error("Failed to load campaign details:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleDispatch = async () => {
    setIsDispatching(true);
    try {
      await dispatchCampaign(campaignId);
      await fetchDetails();
    } catch (err) {
      console.error("Failed to dispatch campaign:", err);
    } finally {
      setIsDispatching(false);
    }
  };

  useEffect(() => {
    if (isOpen && campaignId) {
      fetchDetails();
    }
  }, [isOpen, campaignId]);

  if (!isOpen) return null;

  const recipients = details?.recipients || [];

  const filteredRecipients = recipients.filter((r) => {
    const matchesSearch =
      r.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      r.email.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus =
      statusFilter === "all" || r.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const countSent = recipients.filter((r) => r.status === "sent").length;
  const countFailed = recipients.filter((r) => r.status === "failed").length;
  const countPending = recipients.filter((r) => r.status === "pending").length;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-zinc-950/80 backdrop-blur-md animate-in fade-in duration-200">
      <div className="relative w-full max-w-4xl max-h-[90vh] flex flex-col rounded-2xl border border-zinc-800 bg-zinc-900 shadow-2xl overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-zinc-800 px-6 py-4 bg-zinc-900/50">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-400">
              <Mail className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-base font-bold text-white">
                {details?.title || `Campaign #${campaignId}`}
              </h2>
              <p className="text-xs text-zinc-400 font-mono">
                ID: #{campaignId} &bull; Created {details?.created_at ? new Date(details.created_at).toLocaleDateString() : ""}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            {countPending > 0 && (
              <button
                onClick={handleDispatch}
                disabled={isDispatching}
                className="flex items-center gap-1.5 rounded-xl bg-indigo-600 px-4 py-2 text-xs font-semibold text-white hover:bg-indigo-500 transition-all shadow-md shadow-indigo-600/20 disabled:opacity-50"
              >
                <Send className={`w-3.5 h-3.5 ${isDispatching ? "animate-spin" : ""}`} />
                <span>{isDispatching ? "Enqueuing..." : `Dispatch ${countPending} Pending`}</span>
              </button>
            )}
            <button
              onClick={onClose}
              className="rounded-lg p-1.5 text-zinc-400 hover:bg-zinc-800 hover:text-white transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Content Body */}
        {loading ? (
          <div className="flex h-64 items-center justify-center">
            <RefreshCw className="w-6 h-6 animate-spin text-indigo-400" />
          </div>
        ) : (
          <div className="flex-1 overflow-y-auto p-6 space-y-6">
            {/* Meta Cards */}
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-4">
              <div className="rounded-xl border border-zinc-800 bg-zinc-950/50 p-4">
                <span className="text-[10px] font-semibold uppercase tracking-wider text-zinc-500">
                  Total Recipients
                </span>
                <p className="text-xl font-bold text-white mt-1">
                  {recipients.length}
                </p>
              </div>

              <div className="rounded-xl border border-emerald-500/20 bg-emerald-500/5 p-4">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] font-semibold uppercase tracking-wider text-emerald-400">
                    Dispatched
                  </span>
                  <CheckCircle2 className="w-4 h-4 text-emerald-400" />
                </div>
                <p className="text-xl font-bold text-emerald-400 mt-1">
                  {countSent}
                </p>
              </div>

              <div className="rounded-xl border border-red-500/20 bg-red-500/5 p-4">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] font-semibold uppercase tracking-wider text-red-400">
                    Failed / DLQ
                  </span>
                  <AlertTriangle className="w-4 h-4 text-red-400" />
                </div>
                <p className="text-xl font-bold text-red-400 mt-1">
                  {countFailed}
                </p>
              </div>

              <div className="rounded-xl border border-amber-500/20 bg-amber-500/5 p-4">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] font-semibold uppercase tracking-wider text-amber-400">
                    Pending
                  </span>
                  <Clock className="w-4 h-4 text-amber-400" />
                </div>
                <p className="text-xl font-bold text-amber-400 mt-1">
                  {countPending}
                </p>
              </div>
            </div>

            {/* Template Specs */}
            {details && (
              <div className="rounded-xl border border-zinc-800 bg-zinc-950/50 p-4 space-y-2">
                <div className="flex items-center gap-2 text-xs font-semibold text-zinc-300">
                  <FileText className="w-4 h-4 text-indigo-400" />
                  <span>Subject:</span>
                  <span className="font-normal text-zinc-400">{details.subject}</span>
                </div>
                <div className="text-xs text-zinc-400 bg-zinc-900/80 rounded-lg p-3 font-mono border border-zinc-800/80 whitespace-pre-wrap">
                  {details.body_template}
                </div>
              </div>
            )}

            {/* Filter Bar */}
            <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
              <div className="relative w-full sm:w-72">
                <Search className="absolute left-3 top-2.5 h-4 w-4 text-zinc-500" />
                <input
                  type="text"
                  placeholder="Search recipient name or email..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full rounded-xl border border-zinc-800 bg-zinc-950 pl-9 pr-4 py-2 text-xs text-zinc-200 placeholder-zinc-500 focus:border-indigo-500 focus:outline-none"
                />
              </div>

              <div className="flex items-center gap-1.5 w-full sm:w-auto">
                {["all", "sent", "failed", "pending"].map((status) => (
                  <button
                    key={status}
                    onClick={() => setStatusFilter(status)}
                    className={`rounded-lg px-3 py-1.5 text-xs font-medium capitalize transition-colors ${
                      statusFilter === status
                        ? "bg-indigo-600 text-white"
                        : "bg-zinc-800 text-zinc-400 hover:text-white"
                    }`}
                  >
                    {status}
                  </button>
                ))}
                <button
                  onClick={fetchDetails}
                  className="ml-auto rounded-lg border border-zinc-800 bg-zinc-950 p-2 text-zinc-400 hover:text-white"
                >
                  <RefreshCw className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>

            {/* Recipients Table */}
            <div className="rounded-xl border border-zinc-800 bg-zinc-950 overflow-hidden">
              <table className="w-full text-left text-xs">
                <thead className="border-b border-zinc-800 bg-zinc-900/60 font-semibold text-zinc-400 uppercase tracking-wider text-[10px]">
                  <tr>
                    <th className="px-4 py-3">Recipient</th>
                    <th className="px-4 py-3">Email</th>
                    <th className="px-4 py-3">Status</th>
                    <th className="px-4 py-3">Details / Error</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60 text-zinc-300">
                  {filteredRecipients.length === 0 ? (
                    <tr>
                      <td colSpan={4} className="px-4 py-8 text-center text-zinc-500">
                        No recipients match your search filter.
                      </td>
                    </tr>
                  ) : (
                    filteredRecipients.map((r) => (
                      <tr key={r.id} className="hover:bg-zinc-900/40 transition-colors">
                        <td className="px-4 py-3 font-medium text-white flex items-center gap-2">
                          <User className="w-3.5 h-3.5 text-zinc-500" />
                          <span>{r.name || "N/A"}</span>
                        </td>
                        <td className="px-4 py-3 font-mono text-zinc-400">
                          {r.email}
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wider ${
                              r.status === "sent"
                                ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                                : r.status === "failed"
                                ? "bg-red-500/10 text-red-400 border border-red-500/20"
                                : "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                            }`}
                          >
                            {r.status}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-zinc-400 max-w-xs truncate">
                          {r.error_message ? (
                            <span className="text-red-400 font-mono text-[11px]">
                              {r.error_message}
                            </span>
                          ) : r.sent_at ? (
                            <span className="text-emerald-400/80 font-mono text-[11px]">
                              Sent at {new Date(r.sent_at).toLocaleTimeString()}
                            </span>
                          ) : (
                            <span className="text-zinc-600 font-mono text-[11px]">
                              Queued in Asynq
                            </span>
                          )}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
