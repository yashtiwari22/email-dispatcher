"use client";

import React, { useState, useEffect } from "react";
import { AlertOctagon, X, RotateCw, RefreshCw, Code, CheckCircle2, Loader2 } from "lucide-react";
import { DLQRecord, getDLQRecords, replayDLQRecord } from "@/lib/api";

interface DLQInspectorModalProps {
    isOpen: boolean;
    onClose: () => void;
}

export default function DLQInspectorModal({ isOpen, onClose }: DLQInspectorModalProps) {
    const [records, setRecords] = useState<DLQRecord[]>([]);
    const [loading, setLoading] = useState(false);
    const [replayingId, setReplayingId] = useState<number | null>(null);
    const [expandedPayloadId, setExpandedPayloadId] = useState<number | null>(null);
    const [filter, setFilter] = useState<string>("all");

    const fetchRecords = async () => {
        setLoading(true);
        try {
            const data = await getDLQRecords(filter === "all" ? undefined : filter);
            setRecords(data);
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (isOpen) {
            fetchRecords();
        }
    }, [isOpen, filter]);

    if (!isOpen) return null;

    const handleReplay = async (id: number) => {
        setReplayingId(id);
        try {
            await replayDLQRecord(id);
            setRecords((prev) =>
                prev.map((r) => (r.id === id ? { ...r, status: "replayed" } : r))
            );
        } catch (err) {
            alert("Failed to replay DLQ job");
        } finally {
            setReplayingId(null);
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-md p-4">
            <div className="w-full max-w-4xl rounded-2xl bg-zinc-900 border border-zinc-800 p-6 shadow-2xl transition-all max-h-[85vh] flex flex-col">
                {/* Header */}
                <div className="flex items-center justify-between border-b border-zinc-800 pb-4">
                    <div className="flex items-center gap-3">
                        <div className="rounded-xl bg-red-500/10 p-2.5 text-red-400">
                            <AlertOctagon className="w-6 h-6" />
                        </div>
                        <div>
                            <h3 className="text-xl font-bold text-white">Dead-Letter Queue Inspector</h3>
                            <p className="text-xs text-zinc-400">Inspect & retry failed dispatch jobs</p>
                        </div>
                    </div>
                    <button
                        onClick={onClose}
                        className="rounded-lg p-1.5 text-zinc-400 hover:bg-zinc-800 hover:text-white transition-colors"
                    >
                        <X className="w-5 h-5" />
                    </button>
                </div>

                {/* Filter Controls */}
                <div className="flex items-center justify-between mt-4">
                    <div className="flex gap-2">
                        {["all", "pending", "replayed"].map((f) => (
                            <button
                                key={f}
                                onClick={() => setFilter(f)}
                                className={`rounded-lg px-3 py-1.5 text-xs font-semibold uppercase tracking-wider transition-all ${filter === f
                                        ? "bg-indigo-600 text-white shadow-md shadow-indigo-600/30"
                                        : "bg-zinc-800 text-zinc-400 hover:bg-zinc-700 hover:text-white"
                                    }`}
                            >
                                {f}
                            </button>
                        ))}
                    </div>

                    <button
                        onClick={fetchRecords}
                        disabled={loading}
                        className="flex items-center gap-2 rounded-lg bg-zinc-800 px-3 py-1.5 text-xs font-medium text-zinc-300 hover:bg-zinc-700 hover:text-white transition-colors"
                    >
                        <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin" : ""}`} />
                        <span>Refresh</span>
                    </button>
                </div>

                {/* Table List */}
                <div className="mt-4 flex-1 overflow-y-auto rounded-xl border border-zinc-800 bg-zinc-950/50">
                    {records.length === 0 ? (
                        <div className="p-12 text-center text-zinc-500 text-sm">
                            No DLQ records found for filter <span className="font-semibold text-zinc-400">{filter}</span>.
                        </div>
                    ) : (
                        <table className="w-full text-left text-xs text-zinc-300 border-collapse">
                            <thead className="bg-zinc-900/80 text-zinc-400 font-semibold uppercase border-b border-zinc-800 sticky top-0 backdrop-blur-sm">
                                <tr>
                                    <th className="py-3 px-4">Recipient</th>
                                    <th className="py-3 px-4">Error Cause</th>
                                    <th className="py-3 px-4">Status</th>
                                    <th className="py-3 px-4">Failed At</th>
                                    <th className="py-3 px-4 text-right">Actions</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-zinc-800/60">
                                {records.map((r) => (
                                    <React.Fragment key={r.id}>
                                        <tr className="hover:bg-zinc-900/40 transition-colors">
                                            <td className="py-3 px-4 font-medium text-white">{r.recipient_email}</td>
                                            <td className="py-3 px-4 text-red-400 font-mono max-w-xs truncate">{r.error_reason}</td>
                                            <td className="py-3 px-4">
                                                <span
                                                    className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wider ${r.status === "replayed"
                                                            ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                                                            : "bg-red-500/10 text-red-400 border border-red-500/20"
                                                        }`}
                                                >
                                                    {r.status === "replayed" && <CheckCircle2 className="w-3 h-3" />}
                                                    {r.status}
                                                </span>
                                            </td>
                                            <td className="py-3 px-4 text-zinc-400">
                                                {new Date(r.failed_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                            </td>
                                            <td className="py-3 px-4 text-right">
                                                <div className="flex items-center justify-end gap-2">
                                                    <button
                                                        onClick={() => setExpandedPayloadId(expandedPayloadId === r.id ? null : r.id)}
                                                        className="rounded-lg p-1.5 text-zinc-400 hover:bg-zinc-800 hover:text-white transition-colors"
                                                        title="View Payload JSON"
                                                    >
                                                        <Code className="w-4 h-4" />
                                                    </button>
                                                    {r.status === "pending" && (
                                                        <button
                                                            onClick={() => handleReplay(r.id)}
                                                            disabled={replayingId === r.id}
                                                            className="flex items-center gap-1.5 rounded-lg bg-indigo-600/90 px-3 py-1.5 text-white font-medium hover:bg-indigo-500 transition-all shadow-sm"
                                                        >
                                                            {replayingId === r.id ? (
                                                                <Loader2 className="w-3.5 h-3.5 animate-spin" />
                                                            ) : (
                                                                <RotateCw className="w-3.5 h-3.5" />
                                                            )}
                                                            <span>Replay</span>
                                                        </button>
                                                    )}
                                                </div>
                                            </td>
                                        </tr>
                                        {expandedPayloadId === r.id && (
                                            <tr className="bg-zinc-950">
                                                <td colSpan={5} className="p-4 border-t border-zinc-800">
                                                    <p className="text-[11px] font-semibold text-zinc-400 mb-1">Payload JSON:</p>
                                                    <pre className="rounded-lg bg-black/60 p-3 text-[11px] font-mono text-emerald-400 overflow-x-auto border border-zinc-800">
                                                        {JSON.stringify(JSON.parse(r.payload_json), null, 2)}
                                                    </pre>
                                                </td>
                                            </tr>
                                        )}
                                    </React.Fragment>
                                ))}
                            </tbody>
                        </table>
                    )}
                </div>
            </div>
        </div>
    );
}
