import type React from "react"
import type { StatusDisplayProps } from "../types"

const StatusDisplay: React.FC<StatusDisplayProps> = ({ status, capacity, psnr }) => {
  const statusColors = {
    info: "text-cyan-400 border-cyan-400/50",
    success: "text-green-400 border-green-400/50",
    error: "text-red-400 border-red-400/50",
    warning: "text-yellow-400 border-yellow-400/50",
  }

  const statusIcons = {
    info: "◉",
    success: "✓",
    error: "✗",
    warning: "⚠",
  }

  return (
    <div className="mb-8 max-w-6xl mx-auto">
      {/* Main status display */}
      <div className={`border rounded-lg p-4 bg-black/30 backdrop-blur-sm ${statusColors[status.type]}`}>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2">
            <span className="text-lg animate-pulse">{statusIcons[status.type]}</span>
            <span className="font-mono text-sm">STATUS:</span>
          </div>
          <div className="flex-1">
            {status.isLoading ? (
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 bg-current rounded-full animate-ping"></div>
                <div className="w-2 h-2 bg-current rounded-full animate-ping animation-delay-200"></div>
                <div className="w-2 h-2 bg-current rounded-full animate-ping animation-delay-400"></div>
                <span className="ml-2 terminal-text">{status.message}</span>
              </div>
            ) : (
              <span className="font-mono">{status.message}</span>
            )}
          </div>
        </div>
      </div>

      {/* Additional info panels */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4">
        {/* Capacity display */}
        {capacity && (
          <div className="border border-purple-400/50 rounded-lg p-3 bg-purple-900/20 backdrop-blur-sm">
            <div className="text-purple-400 font-mono text-sm mb-1">CAPACITY</div>
            <div className="text-white">
              <div className="text-xs">{capacity.maxBytes} bytes</div>
              <div className="text-xs">{capacity.maxCharacters} chars</div>
            </div>
          </div>
        )}

        {/* PSNR display */}
        {psnr !== undefined && (
          <div className="border border-green-400/50 rounded-lg p-3 bg-green-900/20 backdrop-blur-sm">
            <div className="text-green-400 font-mono text-sm mb-1">PSNR</div>
            <div className="text-white text-lg font-bold">{psnr.toFixed(2)} dB</div>
          </div>
        )}
      </div>
    </div>
  )
}

export default StatusDisplay
