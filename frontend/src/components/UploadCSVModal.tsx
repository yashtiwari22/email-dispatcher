"use client";

import React, { useState, useRef } from "react";
import { Upload, X, FileText, CheckCircle2, AlertCircle, Loader2 } from "lucide-react";
import { uploadCSV } from "@/lib/api";

interface UploadCSVModalProps {
  campaignId: number;
  campaignTitle: string;
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export default function UploadCSVModal({
  campaignId,
  campaignTitle,
  isOpen,
  onClose,
  onSuccess,
}: UploadCSVModalProps) {
  const [file, setFile] = useState<File | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successResult, setSuccessResult] = useState<{ total_queued: number; invalid_count: number } | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  if (!isOpen) return null;

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = () => {
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const droppedFile = e.dataTransfer.files[0];
      if (droppedFile.name.endsWith(".csv")) {
        setFile(droppedFile);
        setError(null);
      } else {
        setError("Please upload a valid .csv file");
      }
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setFile(e.target.files[0]);
      setError(null);
    }
  };

  const handleUpload = async () => {
    if (!file) return;
    setIsUploading(true);
    setError(null);

    try {
      const res = await uploadCSV(campaignId, file);
      setSuccessResult({
        total_queued: res.total_queued,
        invalid_count: res.invalid_count,
      });
      setTimeout(() => {
        onSuccess();
        onClose();
        setFile(null);
        setSuccessResult(null);
      }, 1500);
    } catch (err: any) {
      setError(err.message || "Failed to upload CSV");
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
      <div className="w-full max-w-lg rounded-2xl bg-zinc-900 border border-zinc-800 p-6 shadow-2xl transition-all">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-zinc-800 pb-4">
          <div>
            <h3 className="text-xl font-bold text-white">Upload Recipients CSV</h3>
            <p className="text-xs text-zinc-400">Target Campaign: <span className="text-indigo-400 font-medium">{campaignTitle}</span></p>
          </div>
          <button
            onClick={onClose}
            className="rounded-lg p-1 text-zinc-400 hover:bg-zinc-800 hover:text-white transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Dropzone */}
        <div className="mt-6">
          {successResult ? (
            <div className="rounded-xl border border-emerald-500/30 bg-emerald-500/10 p-6 text-center">
              <CheckCircle2 className="w-12 h-12 text-emerald-400 mx-auto mb-2 animate-bounce" />
              <h4 className="text-lg font-semibold text-emerald-300">Upload Complete!</h4>
              <p className="text-sm text-zinc-300 mt-1">
                Queued <span className="font-bold text-white">{successResult.total_queued}</span> email tasks.
              </p>
            </div>
          ) : (
            <div
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              onClick={() => fileInputRef.current?.click()}
              className={`flex flex-col items-center justify-center rounded-xl border-2 border-dashed p-8 text-center cursor-pointer transition-all ${
                isDragging
                  ? "border-indigo-500 bg-indigo-500/10"
                  : file
                  ? "border-zinc-700 bg-zinc-800/50"
                  : "border-zinc-800 bg-zinc-900/50 hover:border-zinc-700 hover:bg-zinc-800/30"
              }`}
            >
              <input
                ref={fileInputRef}
                type="file"
                accept=".csv"
                className="hidden"
                onChange={handleFileChange}
              />
              {file ? (
                <div className="flex items-center gap-3 text-indigo-300">
                  <FileText className="w-8 h-8 text-indigo-400" />
                  <div className="text-left">
                    <p className="text-sm font-semibold text-white truncate max-w-[240px]">{file.name}</p>
                    <p className="text-xs text-zinc-400">{(file.size / 1024).toFixed(1)} KB</p>
                  </div>
                </div>
              ) : (
                <>
                  <div className="rounded-full bg-indigo-500/10 p-3 text-indigo-400 mb-3">
                    <Upload className="w-6 h-6" />
                  </div>
                  <p className="text-sm font-medium text-white">Click or drag & drop CSV file</p>
                  <p className="text-xs text-zinc-500 mt-1">Headers must include: <code className="text-indigo-400">email, name</code></p>
                </>
              )}
            </div>
          )}
        </div>

        {error && (
          <div className="mt-4 flex items-center gap-2 rounded-lg border border-red-500/30 bg-red-500/10 p-3 text-xs text-red-400">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {/* Footer */}
        <div className="mt-6 flex justify-end gap-3 border-t border-zinc-800 pt-4">
          <button
            onClick={onClose}
            disabled={isUploading}
            className="rounded-lg px-4 py-2 text-xs font-medium text-zinc-400 hover:bg-zinc-800 hover:text-white transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleUpload}
            disabled={!file || isUploading || !!successResult}
            className="flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-xs font-medium text-white hover:bg-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg shadow-indigo-600/30"
          >
            {isUploading ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                <span>Uploading & Queuing...</span>
              </>
            ) : (
              <span>Start Queue Dispatch</span>
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
