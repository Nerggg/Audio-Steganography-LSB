"use client"

import { useState } from "react"
import Header from "./components/Header"
import EmbedPanel from "./components/EmbedPanel"
import ExtractPanel from "./components/ExtractPanel"
import StatusDisplay from "./components/StatusDisplay"
import type { AppStatus, EmbedResult, ExtractResult, AudioCapacity } from "./types"

function App() {
  const [activeTab, setActiveTab] = useState<"embed" | "extract">("embed")
  const [status, setStatus] = useState<AppStatus>({
    isLoading: false,
    message: "CYBERSTEG TERMINAL READY - SELECT OPERATION MODE",
    type: "info",
  })
  const [capacity, setCapacity] = useState<AudioCapacity | undefined>()
  const [psnr, setPsnr] = useState<number | undefined>()

  const handleStatusUpdate = (newStatus: AppStatus) => {
    setStatus(newStatus)
  }

  const handleCapacityUpdate = (newCapacity: AudioCapacity) => {
    setCapacity(newCapacity)
  }

  const handleEmbedComplete = (result: EmbedResult) => {
    if (result.success && result.psnr) {
      setPsnr(result.psnr)
      setStatus({
        isLoading: false,
        message: `MESSAGE EMBEDDED SUCCESSFULLY! PSNR: ${result.psnr.toFixed(2)} dB`,
        type: "success",
      })
    } else {
      setStatus({
        isLoading: false,
        message: result.message || "EMBEDDING OPERATION FAILED",
        type: "error",
      })
    }
  }

  const handleExtractComplete = (result: ExtractResult) => {
    if (result.success) {
      setStatus({
        isLoading: false,
        message: "MESSAGE EXTRACTION COMPLETED SUCCESSFULLY",
        type: "success",
      })
    } else {
      setStatus({
        isLoading: false,
        message: result.error || "EXTRACTION OPERATION FAILED",
        type: "error",
      })
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-900 via-indigo-900 to-black text-white relative overflow-hidden">
      {/* Matrix rain background effect */}
      <div className="matrix-rain"></div>

      {/* CRT Screen Effect */}
      <div className="fixed inset-0 pointer-events-none z-10">
        <div className="absolute inset-0 bg-gradient-to-b from-transparent via-cyan-500/5 to-transparent animate-pulse"></div>
        <div className="absolute inset-0 bg-gradient-to-r from-transparent via-pink-500/5 to-transparent"></div>
        <div className="scanlines"></div>
      </div>

      {/* Cyberpunk grid overlay */}
      <div className="fixed inset-0 cyber-grid opacity-20 pointer-events-none z-5"></div>

      <div className="relative z-20 container mx-auto px-4 py-8">
        <Header />

        <StatusDisplay status={status} capacity={capacity} psnr={psnr} />

        {/* Tab Navigation with enhanced styling */}
        <div className="flex justify-center mb-8">
          <div className="bg-purple-900/50 rounded-lg p-1 border border-cyan-400/30 backdrop-blur-sm pulse-glow">
            <button
              onClick={() => setActiveTab("embed")}
              className={`px-8 py-4 rounded-md font-bold transition-all duration-300 retro-button ${
                activeTab === "embed"
                  ? "bg-gradient-to-r from-pink-500 to-cyan-400 text-black shadow-lg shadow-pink-500/25 neon-border-pink"
                  : "text-cyan-400 hover:text-white hover:bg-purple-800/50 hover:shadow-lg hover:shadow-cyan-400/25"
              }`}
            >
              <span className="flex items-center gap-2">
                <span className="w-2 h-2 bg-current rounded-full animate-pulse"></span>
                EMBED MESSAGE
              </span>
            </button>
            <button
              onClick={() => setActiveTab("extract")}
              className={`px-8 py-4 rounded-md font-bold transition-all duration-300 retro-button ${
                activeTab === "extract"
                  ? "bg-gradient-to-r from-pink-500 to-cyan-400 text-black shadow-lg shadow-pink-500/25 neon-border-pink"
                  : "text-cyan-400 hover:text-white hover:bg-purple-800/50 hover:shadow-lg hover:shadow-cyan-400/25"
              }`}
            >
              <span className="flex items-center gap-2">
                <span className="w-2 h-2 bg-current rounded-full animate-pulse"></span>
                EXTRACT MESSAGE
              </span>
            </button>
          </div>
        </div>

        {/* Main Content */}
        <div className="max-w-6xl mx-auto">
          {activeTab === "embed" ? (
            <EmbedPanel
              onStatusUpdate={handleStatusUpdate}
              onCapacityUpdate={handleCapacityUpdate}
              onEmbedComplete={handleEmbedComplete}
            />
          ) : (
            <ExtractPanel onStatusUpdate={handleStatusUpdate} onExtractComplete={handleExtractComplete} />
          )}
        </div>

        {/* Footer with system info */}
        <footer className="mt-16 text-center">
          <div className="border-t border-cyan-400/20 pt-8">
            <div className="flex justify-center items-center gap-4 text-sm text-gray-400 font-mono">
              <span className="flex items-center gap-1">
                <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
                SYSTEM ONLINE
              </span>
              <span>|</span>
              <span>ENCRYPTION: AES-256</span>
              <span>|</span>
              <span>LSB: 1-4 BITS</span>
              <span>|</span>
              <span className="hologram">CYBERSTEG v2.1.84</span>
            </div>
          </div>
        </footer>
      </div>
    </div>
  )
}

export default App
