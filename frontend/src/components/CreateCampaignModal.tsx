"use client";

import React, { useState } from "react";
import { PlusCircle, X, Eye, Code, Loader2 } from "lucide-react";
import { createCampaign } from "@/lib/api";

interface CreateCampaignModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export default function CreateCampaignModal({ isOpen, onClose, onSuccess }: CreateCampaignModalProps) {
  const [title, setTitle] = useState("");
  const [subject, setSubject] = useState("");
  const [bodyTemplate, setBodyTemplate] = useState(
    `<div style="font-family: sans-serif; padding: 20px; color: #333;">\n  <h2>Hello {{.Name}},</h2>\n  <p>Welcome to our platform! Your registered email is <strong>{{.Email}}</strong>.</p>\n</div>`
  );
  const [activeTab, setActiveTab] = useState<"editor" | "preview">("editor");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const insertVariable = (varName: string) => {
    setBodyTemplate((prev) => prev + varName);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title || !subject || !bodyTemplate) {
      setError("Please fill out all required fields.");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      await createCampaign({
        title,
        subject,
        body_template: bodyTemplate,
      });
      onSuccess();
      onClose();
      setTitle("");
      setSubject("");
    } catch (err: any) {
      setError(err.message || "Failed to create campaign");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-md p-4">
      <div className="w-full max-w-2xl rounded-2xl bg-zinc-900 border border-zinc-800 p-6 shadow-2xl transition-all">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-zinc-800 pb-4">
          <div className="flex items-center gap-2.5">
            <div className="rounded-xl bg-indigo-500/10 p-2 text-indigo-400">
              <PlusCircle className="w-5 h-5" />
            </div>
            <h3 className="text-xl font-bold text-white">Create New Campaign</h3>
          </div>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-zinc-400 hover:bg-zinc-800 hover:text-white transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="mt-4 space-y-4">
          <div>
            <label className="block text-xs font-semibold text-zinc-400 uppercase tracking-wider mb-1">
              Campaign Title
            </label>
            <input
              type="text"
              placeholder="e.g. Q3 Customer Onboarding"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full rounded-xl bg-zinc-950 border border-zinc-800 px-4 py-2.5 text-sm text-white placeholder-zinc-600 focus:border-indigo-500 focus:outline-none transition-colors"
              required
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-zinc-400 uppercase tracking-wider mb-1">
              Email Subject Line
            </label>
            <input
              type="text"
              placeholder="e.g. Welcome {{.Name}} to Email Dispatcher"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              className="w-full rounded-xl bg-zinc-950 border border-zinc-800 px-4 py-2.5 text-sm text-white placeholder-zinc-600 focus:border-indigo-500 focus:outline-none transition-colors"
              required
            />
          </div>

          <div>
            <div className="flex items-center justify-between mb-1.5">
              <label className="block text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                HTML Body Template
              </label>

              {/* Editor / Preview Tabs */}
              <div className="flex items-center gap-1 rounded-lg bg-zinc-950 p-1 border border-zinc-800">
                <button
                  type="button"
                  onClick={() => setActiveTab("editor")}
                  className={`flex items-center gap-1 rounded-md px-2.5 py-1 text-[11px] font-semibold transition-all ${
                    activeTab === "editor"
                      ? "bg-indigo-600 text-white shadow-sm"
                      : "text-zinc-400 hover:text-white"
                  }`}
                >
                  <Code className="w-3 h-3" />
                  Code
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab("preview")}
                  className={`flex items-center gap-1 rounded-md px-2.5 py-1 text-[11px] font-semibold transition-all ${
                    activeTab === "preview"
                      ? "bg-indigo-600 text-white shadow-sm"
                      : "text-zinc-400 hover:text-white"
                  }`}
                >
                  <Eye className="w-3 h-3" />
                  Live Preview
                </button>
              </div>
            </div>

            {/* Variable Tag Shortcuts */}
            <div className="flex items-center gap-2 mb-2">
              <span className="text-[11px] text-zinc-500 font-medium">Insert tags:</span>
              {["{{.Name}}", "{{.Email}}"].map((tag) => (
                <button
                  key={tag}
                  type="button"
                  onClick={() => insertVariable(tag)}
                  className="rounded-md border border-indigo-500/20 bg-indigo-500/10 px-2 py-0.5 text-[11px] font-mono text-indigo-300 hover:bg-indigo-500/20 transition-colors"
                >
                  + {tag}
                </button>
              ))}
            </div>

            {activeTab === "editor" ? (
              <textarea
                rows={6}
                value={bodyTemplate}
                onChange={(e) => setBodyTemplate(e.target.value)}
                className="w-full rounded-xl bg-zinc-950 border border-zinc-800 p-4 text-xs font-mono text-emerald-400 placeholder-zinc-600 focus:border-indigo-500 focus:outline-none transition-colors"
                required
              />
            ) : (
              <div className="w-full rounded-xl bg-white p-4 min-h-[160px] text-zinc-900 border border-zinc-800 overflow-y-auto">
                <div
                  dangerouslySetInnerHTML={{
                    __html: bodyTemplate
                      .replace(/\{\{\.Name\}\}/g, "Jane Doe")
                      .replace(/\{\{\.Email\}\}/g, "jane@example.com"),
                  }}
                />
              </div>
            )}
          </div>

          {error && <p className="text-xs text-red-400">{error}</p>}

          {/* Footer */}
          <div className="flex justify-end gap-3 border-t border-zinc-800 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="rounded-lg px-4 py-2 text-xs font-medium text-zinc-400 hover:bg-zinc-800 hover:text-white transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-xs font-semibold text-white hover:bg-indigo-500 disabled:opacity-50 transition-all shadow-lg shadow-indigo-600/30"
            >
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Saving...
                </>
              ) : (
                <span>Create Campaign</span>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
